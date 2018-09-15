package main

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	"encoding/base64"
	"encoding/json"
	"k8s.io/api/core/v1"
	"bytes"
	"strconv"
	"io/ioutil"
	"strings"
	"os"
	"github.com/prometheus/common/log"
	"os/exec"
	"runtime"
)

const staged = "F://etc/nginx/nginx.tmpl"
const tmp = "F://etc/nginx/nginx.tmp"
const dest = "F://etc/nginx/nginx.conf"
const check = "nginx -t -c  "
const reload = "nginx -s reload  "
var instance = ""

type NginxItem struct {
	Name 		  string  `json:"name"`
	Port 		  int     `json:"port"`
	ServicePort   int     `json:"servicePort"`
	Type 		  string  `json:"type"`
	Server 		  string  `json:"server"`
	Upstream 	  string  `json:"upstream"`
	Location 	  string  `json:"location"`
	ServerItem    string  `json:"serverItem"`
}

func main() {
	var kubeconfig = "F:/ssl/1/admin.conf";//flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	//flag.Parse()

	instance = "nginx"//os.Getenv("IPS_CONTROLLER_NAME")
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientSet, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	watcher,err := clientSet.CoreV1().Services("").Watch(metav1.ListOptions{LabelSelector:instance+"=1"})

	if err != nil {
		panic(err.Error())
	}

	for {
		select {
		case event, chanOk := <-watcher.ResultChan():
			if chanOk {
				/*services,err:=clientSet.CoreV1().Services("").List(metav1.ListOptions{LabelSelector:"IPS_NG_INSTANCE=nginx"})
				if err!=nil {
					updateNginxConf(services)
				}*/
				if svc,ok:= event.Object.(*v1.Service);ok{
					updateNginx(svc)
				}
			}
		default:
			//fmt.Println("timeout!")
		}
		time.Sleep(2 * time.Second)
	}
}


func updateNginx(service *v1.Service){
	//text := "W3sibG9jYXRpb24iOiJoZWFsdGhfY2hlY2sgaW50ZXJ2YWw9NSBmYWlscz0zIHBhc3Nlcz0yIHVyaT0vc29tZS9wYXRoIiwicG9ydCI6NjY2Niwic2VydmVyIjoic2VuZGZpbGUgICAgICAgICAgICBvbjtcbnRjcF9ub3B1c2ggICAgICAgICAgb247Iiwic2VydmVySXRlbSI6IndlaWdodD01IHNsb3dfc3RhcnQ9MzBzIG1heF9mYWlscz0zIGZhaWxfdGltZW91dD0zMHM7IiwidXBzdHJlYW0iOiJpcF9oYXNoO1xuem9uZSBiYWNrZW5kIDY0aztcbnF1ZXVlIDEwMCB0aW1lb3V0PTcwOyJ9XQ=="
	//按照端口聚合
	data,err:=base64.StdEncoding.DecodeString(service.Annotations["IPS_NG_ITEM"])
	fmt.Println("asdasd",string(data))
	if err!=nil {
		fmt.Println("decode error.")
	}
	items := []NginxItem{}
	json.Unmarshal(data, &items)
	//产生新的配置
	httpBuffer := bytes.NewBufferString("")
	tcpBuffer := bytes.NewBufferString("")
	for _,item := range items {
		fmt.Println("asdasd",item.Name)
		if item.Name != instance {
			continue
		}
		fmt.Println("asdasd",item.Type)
		if item.Type == "HTTP" {
			httpBuffer.WriteString("\nserver {\n")
			httpBuffer.WriteString("  "+item.Server)
			httpBuffer.WriteString("\n  listen  " + strconv.Itoa(item.Port) + ";")
			httpBuffer.WriteString("\n  location / {\n")
			httpBuffer.WriteString("  "+item.Location)
			httpBuffer.WriteString("\n    proxy_pass http://"+service.Name+";")
			httpBuffer.WriteString("\n  }")
			httpBuffer.WriteString("\n}")
			httpBuffer.WriteString("\n   upstream  "+service.Name+" {\n")
			httpBuffer.WriteString("   "+item.Upstream)
			httpBuffer.WriteString("\n   server "+service.Spec.ClusterIP +":"+strconv.Itoa(item.ServicePort)+" " + item.ServerItem )
			httpBuffer.WriteString("\n}")
		}
		if item.Type == "TCP" {
			tcpBuffer.WriteString("\n server {\n")
			tcpBuffer.WriteString("  "+item.Server)
			tcpBuffer.WriteString("\n   listen  " + strconv.Itoa(item.Port) + ";")
			tcpBuffer.WriteString("\n   proxy_pass "+service.Name+";")
			tcpBuffer.WriteString("\n }")
			tcpBuffer.WriteString("\n   upstream  "+service.Name+" {")
			tcpBuffer.WriteString("\n      "+item.Upstream)
			tcpBuffer.WriteString("\n   server "+service.Spec.ClusterIP +":"+strconv.Itoa(item.ServicePort)+" " + item.ServerItem )
			tcpBuffer.WriteString("\n }")
		}

	}
	b, err := ioutil.ReadFile(staged)
	buffer := bytes.NewBuffer(b)

	fmt.Println("000000",buffer.String())
	httpContent := strings.Replace(buffer.String(), "#HTTP_SERVER#", httpBuffer.String(), -1)

	tcpContent := strings.Replace(httpContent, "#TCP_SERVER#", tcpBuffer.String(), -1)

	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err!=nil {
		if strings.Contains(err.Error(), "The system cannot find the path specified") {
			f,_= os.Create(tmp)
			log.Error(fmt.Sprintf("%s",err))
		}
	}
	defer f.Close()
	fmt.Println(tcpContent)
	f.WriteString(tcpContent)

	//err = runCommand(check + tmp)
	//
	//if err == nil {
	//	os.Rename(tmp,dest)
	//	runCommand(reload + dest)
	//}
}


func runCommand(cmd string) error {
	log.Debug("Running " + cmd)
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("/bin/sh", "-c", cmd)
	}

	output, err := c.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	log.Debug(fmt.Sprintf("%q", string(output)))
	return nil
}



func updateNginxConf(services *v1.ServiceList){
	//text := "W3sibG9jYXRpb24iOiJoZWFsdGhfY2hlY2sgaW50ZXJ2YWw9NSBmYWlscz0zIHBhc3Nlcz0yIHVyaT0vc29tZS9wYXRoIiwicG9ydCI6NjY2Niwic2VydmVyIjoic2VuZGZpbGUgICAgICAgICAgICBvbjtcbnRjcF9ub3B1c2ggICAgICAgICAgb247Iiwic2VydmVySXRlbSI6IndlaWdodD01IHNsb3dfc3RhcnQ9MzBzIG1heF9mYWlscz0zIGZhaWxfdGltZW91dD0zMHM7IiwidXBzdHJlYW0iOiJpcF9oYXNoO1xuem9uZSBiYWNrZW5kIDY0aztcbnF1ZXVlIDEwMCB0aW1lb3V0PTcwOyJ9XQ=="

	group := make(map[int][]NginxItem)

	//按照端口聚合
	for _,sitem := range services.Items {
		data,err:=base64.StdEncoding.DecodeString(sitem.Annotations["IPS_NG_ITEM"])
		if err!=nil {
			fmt.Println("decode error.")
		}
		items := []NginxItem{}
		json.Unmarshal(data, &items)
		for _,item := range items {
			item.ServerItem = sitem.Spec.ClusterIP +" " + item.ServerItem
			if _, ok := group[item.Port]; ok {
				group[item.Port] = append(group[item.Port], item)
			}else{
				group[item.Port] = []NginxItem{item}
			}
		}
	}
	//产生新的配置
	httpBuffer := bytes.NewBufferString("")
	for key,items := range group {
		httpBuffer.WriteString("server {")
		for index,item := range items {
			if index==0 {
				httpBuffer.WriteString("server {\n")
				httpBuffer.WriteString("\tlisten\t\t"+strconv.Itoa(item.Port)+";")
				httpBuffer.WriteString("\t\tproxy_pass http://backend;")
				httpBuffer.WriteString(item.Location)
				httpBuffer.WriteString("\tlocation / {" )
				httpBuffer.WriteString("\t}" )
				httpBuffer.WriteString("upstream  backend {" )
				httpBuffer.WriteString(item.Upstream )
			}
			httpBuffer.WriteString("server "+item.ServerItem)
		}
		httpBuffer.WriteString("}" )
		httpBuffer.WriteString("}")
		fmt.Println(key,items)
	}


}




func rr(){
	items := []NginxItem{}
	text := "W3sibG9jYXRpb24iOiJoZWFsdGhfY2hlY2sgaW50ZXJ2YWw9NSBmYWlscz0zIHBhc3Nlcz0yIHVyaT0vc29tZS9wYXRoIiwicG9ydCI6NjY2Niwic2VydmVyIjoic2VuZGZpbGUgICAgICAgICAgICBvbjtcbnRjcF9ub3B1c2ggICAgICAgICAgb247Iiwic2VydmVySXRlbSI6IndlaWdodD01IHNsb3dfc3RhcnQ9MzBzIG1heF9mYWlscz0zIGZhaWxfdGltZW91dD0zMHM7IiwidXBzdHJlYW0iOiJpcF9oYXNoO1xuem9uZSBiYWNrZW5kIDY0aztcbnF1ZXVlIDEwMCB0aW1lb3V0PTcwOyJ9XQ=="
	data,_:=base64.StdEncoding.DecodeString(text)
	json.Unmarshal(data, &items)

	fmt.Println(len(items))
	fmt.Println(items[0].Location)
}

