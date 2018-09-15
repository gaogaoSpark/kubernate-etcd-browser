package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"github.com/golang/glog"
	"bytes"
)

func main1() {

	crt, err := tls.LoadX509KeyPair("./ssl/etcd.pem", "./ssl/etcd-key.pem")
	if err != nil {
		fmt.Println(err.Error())
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	tlsConfig.Certificates = []tls.Certificate{crt}
	if caCert, err := ioutil.ReadFile("./ssl/ca.pem"); err != nil {
		glog.Errorf("failed to read ca file while getting backends: %s", err)
	} else {
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caPool
		tlsConfig.InsecureSkipVerify = false
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"10.20.16.227:2379", "10.20.16.228:2379", "10.20.16.229:2379"},
		DialTimeout: 5 * time.Second,
		TLS: tlsConfig,
	})

	if err != nil {
		fmt.Println("err")
	}
	ctx := cli.Ctx()

	res,err1 := cli.KV.Get(ctx,"/",clientv3.WithFromKey(),clientv3.WithPrefix())
	first := res.Kvs[253]
	//fmt.Println(len(res.Kvs))
	if err1 != nil {
		fmt.Println("err1")
	}else{
		/*for _, kv:= range res.Kvs {
			fmt.Println("-----------------------------------")
			fmt.Println(string(bytes.TrimPrefix(kv.Value ,[]byte(string(kv.Key)))))
		}*/
		/*fmt.Println(string(first.Key))*/
		fmt.Println(string(first.Key))
		fmt.Println("-----------------------------------")

		dd := bytes.HasPrefix(first.Value ,first.Key)
		fmt.Println("-----------------------------------",dd)
		fmt.Println(string(first.Value[0:10]))
		fmt.Println("-----------------------------------")
		fmt.Println(string(first.Value[0:20]))
		fmt.Println("-----------------------------------")
		fmt.Println(string(bytes.TrimPrefix(first.Value ,first.Key)))
	}
	//key作为[]byte(string(d))

	defer cli.Close()


/*	var Scheme = runtime.NewScheme()
	var Codecs = serializer.NewCodecFactory(Scheme)
	s := this.Serializer
	decoders := []runtime.Decoder{s, Codecs.UniversalDeserializer()}

	gv := schema.GroupVersion{"service", "v1"}
	decoder := Codecs.DecoderToVersion(
		recognizer.NewDecoder(decoders...),
		runtime.NewMultiGroupVersioner(
			gv,
			schema.GroupKind{Group: opts.MemoryVersion.Group},
			schema.GroupKind{Group: opts.StorageVersion.Group},
		),
	)


	// Codecs provides access to encoding and decoding for the scheme
	decoder.Decode()
*/
/*	var transformer = etcd3.IdentityTransformer
	transformer.TransformFromStorage()*/

	/*// Codecs provides access to encoding and decoding for the scheme
	var Codecs = serializer.NewCodecFactory(Scheme)
	_, _, err := codec.Decode(value, nil, objPtr)
	var*/
}
