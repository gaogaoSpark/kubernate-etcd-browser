package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/version"
	"net/http"
	"os"
)

var kubeClient *kubernetes.Clientset = nil
var master string = ""
var groupVersion string = ""

type TreeNode struct{
	Label string `json:"label"`
	Data  interface{} `json:"data"`
}
func main() {

/*	var kubeconfig = "F:/ssl/1/admin.conf";//

	if(kubeconfig == ""){
		fmt.Println("kubeconfig is required.")
		os.Exit(0);
	}*/

	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	if(*kubeconfig == ""){
		fmt.Println("kubeconfig is required.")
		os.Exit(0);
	}
	
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	master = config.Host
	if err != nil {
		panic(err.Error())
	}

	client, err := kubernetes.NewForConfig(config)

	groupVersion = client.RESTClient().APIVersion().Version
	if err != nil {
		panic(err.Error())
	}

	kubeClient = client
	http.HandleFunc("/", filter)

	h := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", h))
	err = http.ListenAndServe(":8080", nil) //设置监听的端口
	if err != nil {
		fmt.Println("ListenAndServe error")
	}
}

func filter(w http.ResponseWriter, r *http.Request){
	fmt.Println(r.URL)
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Add("Access-Control-Allow-Headers","Content-Type")//header的类型
		//w.Header().Set("content-type","application/json")//返回数据格式是json
	}
	if r.Method == "OPTIONS" {
		return
	}
	switch r.RequestURI {
		case "/cluster": cluster(w,r)
			break
		case "/namespace": namespace(w,r)
			break
		case "/deployment": deployment(w,r)
			break;
		case "/daemonset": daemonset(w,r)
			break
		case "/pod": pod(w,r)
			break
		case "/event": event(w,r)
			break
		case "/service": service(w,r)
			break
		case "/ingress": ingress(w,r)
			break
		case "/secret": secret(w,r)
			break
		case "/node": node(w,r)
			break
		case "/configMap": configMap(w,r)
			break;
		case "/endpoint": endpoint(w,r)
			break
		case "/serviceAccount": serviceAccount(w,r)
			break
		case "/persistentVolume": persistentVolume(w,r)
			break
		case "/persistentVolumeClaim": persistentVolumeClaim(w,r)
			break
		case "/componentStatus": componentStatus(w,r)
			break
		case "/resourceQuota": resourceQuota(w,r)
			break
		case "/podTemplate": podTemplate(w,r)
			break
		case "/limitRange": limitRange(w,r)
			break
		case "/replicationController": replicationController(w,r)
			break
		case "/podSecurityPolicy": podSecurityPolicy(w,r)
			break
		case "/replicaSet": replicaSet(w,r)
			break
		default:
			http.Redirect(w, r, "/static/index.html", http.StatusFound)
	}

}


func cluster(w http.ResponseWriter, r *http.Request) { //返回数据格式是json
	infos := []string{}
	infos = append(infos,"Kubernetes master is running at "+master)
	services,_:= kubeClient.CoreV1().Services("").List(metav1.ListOptions{})

	for _,svc := range services.Items{
		if svc.Name == "eventer" {
			infos = append(infos,"Heapster is running at "+master+"/api/"+groupVersion+"/namespaces/kube-system/services/eventer/proxy")
			fmt.Println(infos[1])
		}
		if svc.Name == "heapster" {
			infos = append(infos,"Heapster is running at "+master+"/api/v1/namespaces/kube-system/services/heapster/proxy")

		}
	}
	serverVersion,_:= kubeClient.ServerVersion()
	clientVersion := version.Get()
	result := struct {
		ServerVersion interface{}
		ClientVersion interface{}
		Information []string
	}{
		serverVersion,
		clientVersion,
		infos,
	}

	data,_ := json.Marshal(result)
	fmt.Fprintf(w, string(data))
}

func namespace(w http.ResponseWriter, r *http.Request) { //返回数据格式是json
	namespaces,err:= kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get namespace.")
	}
	result := []TreeNode{}
	for _,ns := range namespaces.Items{
		node := TreeNode{}
		node.Label = ns.Name
		node.Data = ns
		result = append(result, node)
	}
	data,_ := json.Marshal(result)
	fmt.Fprintf(w, string(data))
}

func deployment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	deployments,err := kubeClient.ExtensionsV1beta1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Deployments.")
	}
	data,_ := json.Marshal(deployments.Items)
	w.Write([]byte(string(data)))
	//直接输出字符串有问题
	//fmt.Fprintf(w, string(data))
}

func daemonset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	daemonsets,err := kubeClient.ExtensionsV1beta1().DaemonSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Daemonsets.")
	}
	data,_ := json.Marshal(daemonsets.Items)
	w.Write([]byte(string(data)))
}

func pod(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	pods,err := kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Pods.")
	}
	data,_ := json.Marshal(pods.Items)
	w.Write([]byte(string(data)))
}

func service(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	services,err := kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Services.")
	}
	data,_ := json.Marshal(services.Items)
	w.Write([]byte(string(data)))
}

func ingress(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	ingresses,err := kubeClient.ExtensionsV1beta1().Ingresses(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Ingress.")
	}
	data,_ := json.Marshal(ingresses.Items)
	w.Write([]byte(string(data)))
}

func node(w http.ResponseWriter, r *http.Request) {
	nodes,err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Nodes.")
	}
	data,_ := json.Marshal(nodes.Items)
	w.Write([]byte(string(data)))
}

func secret(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	secrets,err := kubeClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Secrets.")
	}
	data,_ := json.Marshal(secrets.Items)
	w.Write([]byte(string(data)))
}

func event(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	events,err := kubeClient.CoreV1().Events(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Events.")
	}
	data,_ := json.Marshal(events.Items)
	w.Write([]byte(string(data)))
}

func configMap(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	configMaps,err := kubeClient.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get ConfigMaps.")
	}
	data,_ := json.Marshal(configMaps.Items)
	w.Write([]byte(string(data)))
}

func endpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	endpoints,err := kubeClient.CoreV1().Endpoints(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get Endpoints.")
	}
	data,_ := json.Marshal(endpoints.Items)
	w.Write([]byte(string(data)))
}

func serviceAccount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	serviceAccounts,err := kubeClient.CoreV1().ServiceAccounts(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get ServiceAccounts.")
	}
	data,_ := json.Marshal(serviceAccounts.Items)
	w.Write([]byte(string(data)))
}

func persistentVolume(w http.ResponseWriter, r *http.Request) {
	persistentVolumes,err := kubeClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get PersistentVolumes.")
	}
	data,_ := json.Marshal(persistentVolumes.Items)
	w.Write([]byte(string(data)))
}

func persistentVolumeClaim(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	persistentVolumeClaims,err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get PersistentVolumeClaims.")
	}
	data,_ := json.Marshal(persistentVolumeClaims.Items)
	w.Write([]byte(string(data)))
}

func componentStatus(w http.ResponseWriter, r *http.Request) {
	componentStatuses,err := kubeClient.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get ComponentStatus.")
	}
	data,_ := json.Marshal(componentStatuses.Items)
	w.Write([]byte(string(data)))
}

func resourceQuota(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	resourceQuotas,err := kubeClient.CoreV1().ResourceQuotas(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get ResourceQuotas.")
	}
	data,_ := json.Marshal(resourceQuotas.Items)
	w.Write([]byte(string(data)))
}

func podTemplate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	podTemplates,err := kubeClient.CoreV1().PodTemplates(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get PodTemplates.")
	}
	data,_ := json.Marshal(podTemplates.Items)
	w.Write([]byte(string(data)))
}

func limitRange(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	limitRanges,err := kubeClient.CoreV1().LimitRanges(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get LimitRanges.")
	}
	data,_ := json.Marshal(limitRanges.Items)
	w.Write([]byte(string(data)))
}

func replicationController(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	replicationController,err := kubeClient.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get ReplicationControllers.")
	}
	data,_ := json.Marshal(replicationController.Items)
	w.Write([]byte(string(data)))
}

func podSecurityPolicy(w http.ResponseWriter, r *http.Request) {
	podSecurityPolicies,err := kubeClient.ExtensionsV1beta1().PodSecurityPolicies().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get PodSecurityPolicies.")
	}
	data,_ := json.Marshal(podSecurityPolicies.Items)
	w.Write([]byte(string(data)))
}

func replicaSet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	replicaSets,err := kubeClient.ExtensionsV1beta1().ReplicaSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("error to get ReplicaSets.")
	}
	data,_ := json.Marshal(replicaSets.Items)
	w.Write([]byte(string(data)))
}

/*func Scales(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	queryForm, err := ioutil.ReadAll(r.Body)
	namespace := "default"
	var params map[string]string
	if err :=json.Unmarshal([]byte(string(queryForm)),&params);err==nil{
		namespace = params["namespace"]
	}else{
		fmt.Println(err)
	}
	replicaSets,err := kubeClient.ExtensionsV1beta1().Scales(namespace).
	if err != nil {
		fmt.Println("error to get ReplicaSets.")
	}
	data,_ := json.Marshal(replicaSets.Items)
	w.Write([]byte(string(data)))
}*/