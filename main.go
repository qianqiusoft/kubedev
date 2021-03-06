package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
	models "github.com/kubedev/models"
	"github.com/kubedev/utils"
	"github.com/relferreira/sse"
	cors "github.com/rs/cors/wrapper/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	r := gin.Default()
	var kubeconfig = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	r.Use(cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:1234"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Authorization", "Accept", "Origin", "Content-Type"},
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              true,
	}))

	box := packr.NewBox("./dist")
	// r.StaticFS("/", box)
	r.Use(utils.Serve("/", box))
	r.NoRoute(utils.RedirectIndex())

	r.GET("/api", func(c *gin.Context) {
		namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, namespaces)
	})

	r.GET("/api/:namespace/search", func(c *gin.Context) {

		services, _ := clientset.CoreV1().Services(metav1.NamespaceAll).List(metav1.ListOptions{})
		deployments, _ := clientset.AppsV1beta2().Deployments(metav1.NamespaceAll).List(metav1.ListOptions{})
		pods, _ := clientset.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
		cronJobs, _ := clientset.BatchV1beta1().CronJobs(metav1.NamespaceAll).List(metav1.ListOptions{})
		jobs, _ := clientset.BatchV1().Jobs(metav1.NamespaceAll).List(metav1.ListOptions{})

		response := models.SearchResponse{Services: services, Deployments: deployments, Pods: pods, CronJobs: cronJobs, Jobs: jobs}
		c.JSON(200, response)
	})

	r.GET("/api/:namespace/services", func(c *gin.Context) {
		namespace := c.Param("namespace")

		services, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})

		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, services)
	})

	r.GET("/api/:namespace/services/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		service, err := clientset.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, service)
	})

	r.DELETE("/api/:namespace/services/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		police := metav1.DeletePropagationForeground
		deleteOptions := metav1.DeleteOptions{PropagationPolicy: &police}
		err := clientset.CoreV1().Services(namespace).Delete(name, &deleteOptions)
		if err != nil {
			panic(err.Error())
		}

		c.Status(200)
	})

	r.GET("/api/:namespace/deployments", func(c *gin.Context) {
		namespace := c.Param("namespace")

		deployments, err := clientset.AppsV1beta2().Deployments(namespace).List(metav1.ListOptions{})

		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, deployments)
	})

	r.GET("/api/:namespace/deployments/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		deployment, err := clientset.AppsV1beta2().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		selector, errSelector := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
		if errSelector != nil {
			panic(errSelector.Error())
		}

		pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			panic(err.Error())
		}

		response := models.DeploymentResponse{Deployment: deployment, Pods: pods}
		c.JSON(200, response)
	})

	r.DELETE("/api/:namespace/deployments/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		police := metav1.DeletePropagationForeground
		deleteOptions := metav1.DeleteOptions{PropagationPolicy: &police}
		err := clientset.AppsV1beta2().Deployments(namespace).Delete(name, &deleteOptions)
		if err != nil {
			panic(err.Error())
		}

		c.Status(200)
	})

	r.POST("/api/:namespace/deployments/:name/scale", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		var scaleCmd models.ScaleCommand
		err := c.BindJSON(&scaleCmd)

		deployment, err := clientset.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		*deployment.Spec.Replicas = *scaleCmd.Scale
		newDeployment, newErr := clientset.AppsV1().Deployments(namespace).Update(deployment)
		if newErr != nil {
			panic(newErr.Error())
		}

		c.JSON(200, newDeployment)
	})

	r.GET("/api/:namespace/jobs", func(c *gin.Context) {
		namespace := c.Param("namespace")

		jobs, err := clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, jobs)
	})

	r.GET("/api/:namespace/jobs/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		job, err := clientset.BatchV1().Jobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, job)
	})

	r.DELETE("/api/:namespace/jobs/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		police := metav1.DeletePropagationForeground
		deleteOptions := metav1.DeleteOptions{PropagationPolicy: &police}
		err := clientset.BatchV1().Jobs(namespace).Delete(name, &deleteOptions)
		if err != nil {
			panic(err.Error())
		}

		c.Status(200)
	})

	r.GET("/api/:namespace/cron-jobs", func(c *gin.Context) {
		namespace := c.Param("namespace")

		cronJobs, err := clientset.BatchV1beta1().CronJobs(namespace).List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, cronJobs)
	})

	r.GET("/api/:namespace/cron-jobs/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		cronJob, err := clientset.BatchV1beta1().CronJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, cronJob)
	})

	r.DELETE("/api/:namespace/cron-jobs/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		police := metav1.DeletePropagationForeground
		deleteOptions := metav1.DeleteOptions{PropagationPolicy: &police}
		err := clientset.BatchV1beta1().CronJobs(namespace).Delete(name, &deleteOptions)
		if err != nil {
			panic(err.Error())
		}

		c.Status(200)
	})

	r.POST("/api/:namespace/cron-jobs/:name/schedule", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		var scheduleCmd models.ScheduleCommand
		err := c.BindJSON(&scheduleCmd)

		cronJob, err := clientset.BatchV1beta1().CronJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		cronJob.Spec.Schedule = *scheduleCmd.Schedule
		newCronJob, updateErr := clientset.BatchV1beta1().CronJobs(namespace).Update(cronJob)
		if updateErr != nil {
			panic(updateErr.Error())
		}

		c.JSON(200, newCronJob)
	})

	r.GET("/api/:namespace/pods", func(c *gin.Context) {
		namespace := c.Param("namespace")

		pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		// data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw()
		// if err != nil {
		// 	panic(err.Error())
		// }

		// log.Println(string(data))

		c.JSON(200, pods)
	})

	r.GET("/api/:namespace/pods/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		pod, err := clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		c.JSON(200, pod)
	})

	r.DELETE("/api/:namespace/pods/:name", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		police := metav1.DeletePropagationForeground
		deleteOptions := metav1.DeleteOptions{PropagationPolicy: &police}
		err := clientset.CoreV1().Pods(namespace).Delete(name, &deleteOptions)
		if err != nil {
			panic(err.Error())
		}

		c.Status(200)
	})

	r.GET("/api/:namespace/pods/:name/:container/logs", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		logOptions := v1.PodLogOptions{}
		req := clientset.CoreV1().Pods(namespace).GetLogs(name, &logOptions)
		podLogs, err := req.Stream()
		if err != nil {
			panic(err.Error())
		}

		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			panic("error in copy information from podLogs to buf")
		}
		str := buf.String()
		c.JSON(200, gin.H{
			"log": str,
		})
	})

	r.GET("/api/:namespace/pods/:name/:container/logs/stream", func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		container := c.Param("container")
		logOptions := v1.PodLogOptions{
			Container: container,
			Follow:    true,
		}
		req := clientset.CoreV1().Pods(namespace).GetLogs(name, &logOptions)
		podLogs, err := req.Stream()
		if err != nil {
			panic(err.Error())
		}

		defer podLogs.Close()
		chanStream := make(chan string)
		go func() {
			defer close(chanStream)
			reader := bufio.NewReader(podLogs)
			for {
				line, _ := reader.ReadBytes('\n')
				chanStream <- string(line)
			}
		}()
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-chanStream; ok {
				c.Render(-1, sse.Event{
					Data: map[string]interface{}{
						"date":    time.Now().Unix(),
						"content": msg,
					},
				})
				return true
			}
			return false
		})
	})

	r.Run(":9898") // listen and serve on 0.0.0.0:8080
}
