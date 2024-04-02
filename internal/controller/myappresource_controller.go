// controllers/myappresource_controller.go

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	mohanbinarybutterv1alpha1 "github.com/mohanpedala/deployer/api/v1alpha1"
)

// MyAppResourceReconciler reconciles a MyAppResource object
type MyAppResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=my.api.group,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.api.group,resources=myappresources/status,verbs=get;update;patch

func (r *MyAppResourceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := r.Log.WithValues("myappresource", req.NamespacedName)

	// Fetch the MyAppResource instance
	myAppResource := &mohanbinarybutterv1alpha1.MyAppResource{}
	err := r.Get(ctx, req.NamespacedName, myAppResource)
	if err != nil {
		log.Error(err, "Failed to get MyAppResource")
		return reconcile.Result{}, err
	}

	// Define a Deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: myAppResource.Namespace,
			Name:      myAppResource.Name + "-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &myAppResource.Spec.ReplicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": myAppResource.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": myAppResource.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "podinfo",
							Image: myAppResource.Spec.Image.Repository + ":" + myAppResource.Spec.Image.Tag,
							Env: []corev1.EnvVar{
								{
									Name:  "PODINFO_CACHE_SERVER",
									Value: "tcp://" + myAppResource.Name + "-redis" + ":" + "6379/echo",
								},
								{
									Name:  "PODINFO_UI_COLOR",
									Value: myAppResource.Spec.UI.Color,
								},
								{
									Name:  "PODINFO_UI_MESSAGE",
									Value: myAppResource.Spec.UI.Message,
								},
							},
						},
					},
				},
			},
		},
	}

	// Set MyAppResource instance as the owner and controller
	if err := controllerutil.SetControllerReference(myAppResource, deployment, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the Deployment already exists
	foundDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{Namespace: deployment.Namespace, Name: deployment.Name}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Deployment already exists - don't requeue
	log.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)

	// Define a Service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: myAppResource.Namespace,
			Name:      myAppResource.Name + "-service",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": myAppResource.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "port1",
					Protocol:   corev1.ProtocolTCP,
					Port:       9898,
					TargetPort: intstr.FromInt(9898), // Update with the actual port
				},
				{
					Name:       "port2",
					Protocol:   corev1.ProtocolTCP,
					Port:       9999,
					TargetPort: intstr.FromInt(9999), // Update with the actual port
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Set MyAppResource instance as the owner and controller
	if err := controllerutil.SetControllerReference(myAppResource, service, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the Service already exists
	foundService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Namespace: service.Namespace, Name: service.Name}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	// Service already exists - don't requeue
	log.Info("Skip reconcile: Service already exists", "Service.Namespace", foundService.Namespace, "Service.Name", foundService.Name)

	// Check if Redis is enabled in the spec
	// redis connection pod

	if myAppResource.Spec.Redis.Enabled {
		// Define a Deployment object for Redis
		redisDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: myAppResource.Namespace,
				Name:      myAppResource.Name + "-cache-info-redis",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "cache-info-redis",
					},
				},
				Replicas: &myAppResource.Spec.ReplicaCount,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "cache-info-redis",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "redis",
								Image: "redis:latest",
								Ports: []corev1.ContainerPort{
									{
										Name:          "redis",
										ContainerPort: 6379,
										Protocol:      corev1.ProtocolTCP,
									},
								},
								Command: []string{
									"redis-server",
									"redis-master/redis.conf",
								},
							},
						},
					},
				},
			},
		}

		// Set MyAppResource instance as the owner and controller
		if err := controllerutil.SetControllerReference(myAppResource, redisDeployment, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if the Deployment already exists
		foundRedisDeployment := &appsv1.Deployment{}
		err = r.Get(ctx, client.ObjectKey{Namespace: redisDeployment.Namespace, Name: redisDeployment.Name}, foundRedisDeployment)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating a new Deployment for Redis", "Deployment.Namespace", redisDeployment.Namespace, "Deployment.Name", redisDeployment.Name)
			err = r.Create(ctx, redisDeployment)
			if err != nil {
				log.Error(err, "Failed to create new Deployment for Redis", "Deployment.Namespace", redisDeployment.Namespace, "Deployment.Name", redisDeployment.Name)
				return reconcile.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment for Redis")
			return reconcile.Result{}, err
		}

		// Deployment already exists - don't requeue
		log.Info("Skip reconcile: Deployment for Redis already exists", "Deployment.Namespace", foundRedisDeployment.Namespace, "Deployment.Name", foundRedisDeployment.Name)
	}

	// Define a Service object for Redis tcp connection
	redisService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: myAppResource.Namespace,
			Name:      myAppResource.Name + "-redis",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "cache-info-redis",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Protocol:   corev1.ProtocolTCP,
					Port:       6379,
					TargetPort: intstr.FromInt(6379),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	// Set MyAppResource instance as the owner and controller
	if err := controllerutil.SetControllerReference(myAppResource, redisService, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the Service already exists
	foundRedisService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Namespace: redisService.Namespace, Name: redisService.Name}, foundRedisService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service for Redis", "Service.Namespace", redisService.Namespace, "Service.Name", redisService.Name)
		err = r.Create(ctx, redisService)
		if err != nil {
			log.Error(err, "Failed to create new Service for Redis", "Service.Namespace", redisService.Namespace, "Service.Name", redisService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service for Redis")
		return reconcile.Result{}, err
	}

	// Service already exists - don't requeue
	log.Info("Skip reconcile: Service for Redis already exists", "Service.Namespace", foundRedisService.Namespace, "Service.Name", foundRedisService.Name)

	// redis server info
	if myAppResource.Spec.Redis.Enabled {
		// Define a Deployment object for Redis
		redisDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: myAppResource.Namespace,
				Name:      myAppResource.Name + "-podinfo-redis",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "podinfo-redis",
					},
				},
				Replicas: &myAppResource.Spec.ReplicaCount,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "podinfo-redis",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "redis",
								Image: "redis:latest",
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										ContainerPort: 9898,
										Protocol:      corev1.ProtocolTCP,
									},
									{
										Name:          "http-metrics",
										ContainerPort: 9797,
										Protocol:      corev1.ProtocolTCP,
									},
									{
										Name:          "grpc",
										ContainerPort: 9999,
										Protocol:      corev1.ProtocolTCP,
									},
								},
								Command: []string{
									"./podinfo",
									"--port=9898",
									"--cert-path=/data/cert",
									"--port-metrics=9797",
									"--grpc-port=9999",
									"--grpc-service-name=podinfo",
									"--cache-server=tcp://backend-podinfo-redis:6379",
									"--level=info",
									"--random-delay=false",
									"--random-error=false",
								},
							},
						},
					},
				},
			},
		}

		// Set MyAppResource instance as the owner and controller
		if err := controllerutil.SetControllerReference(myAppResource, redisDeployment, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if the Deployment already exists
		foundRedisDeployment := &appsv1.Deployment{}
		err = r.Get(ctx, client.ObjectKey{Namespace: redisDeployment.Namespace, Name: redisDeployment.Name}, foundRedisDeployment)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating a new Deployment for Redis", "Deployment.Namespace", redisDeployment.Namespace, "Deployment.Name", redisDeployment.Name)
			err = r.Create(ctx, redisDeployment)
			if err != nil {
				log.Error(err, "Failed to create new Deployment for Redis", "Deployment.Namespace", redisDeployment.Namespace, "Deployment.Name", redisDeployment.Name)
				return reconcile.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment for Redis")
			return reconcile.Result{}, err
		}

		// Deployment already exists - don't requeue
		log.Info("Skip reconcile: Deployment for Redis already exists", "Deployment.Namespace", foundRedisDeployment.Namespace, "Deployment.Name", foundRedisDeployment.Name)
	}

	// Define a Service object for Redis server pod
	redisPodInfoService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: myAppResource.Namespace,
			Name:      myAppResource.Name + "podinfo-redis",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "podinfo-redis",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "port1",
					Protocol:   corev1.ProtocolTCP,
					Port:       9898,
					TargetPort: intstr.FromInt(9898),
				},
				{
					Name:       "port2",
					Protocol:   corev1.ProtocolTCP,
					Port:       9999,
					TargetPort: intstr.FromInt(9999),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Set MyAppResource instance as the owner and controller
	if err := controllerutil.SetControllerReference(myAppResource, redisPodInfoService, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the Service already exists
	foundRedisPodInfoService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Namespace: redisPodInfoService.Namespace, Name: redisPodInfoService.Name}, foundRedisPodInfoService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service for Redis Pod Info", "Service.Namespace", redisPodInfoService.Namespace, "Service.Name", redisPodInfoService.Name)
		err = r.Create(ctx, redisPodInfoService)
		if err != nil {
			log.Error(err, "Failed to create new Service for Redis Pod Info", "Service.Namespace", redisPodInfoService.Namespace, "Service.Name", redisPodInfoService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service for Redis Pod Info")
		return reconcile.Result{}, err
	}

	// Service already exists - don't requeue
	log.Info("Skip reconcile: Service for Redis Pod Info already exists", "Service.Namespace", foundRedisPodInfoService.Namespace, "Service.Name", foundRedisPodInfoService.Name)

	return reconcile.Result{}, nil
}

func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, err := ctrl.NewControllerManagedBy(mgr).
		For(&mohanbinarybutterv1alpha1.MyAppResource{}).
		Build(r)
	if err != nil {
		return err
	}
	return nil
}
