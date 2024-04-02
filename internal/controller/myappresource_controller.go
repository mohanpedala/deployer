// controllers/myappresource_controller.go

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
									Value: "tcp://<host>:<port>", // Update with your values
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

	// Check if Redis is enabled in the spec
	if myAppResource.Spec.Redis.Enabled {
		// Define a Deployment object for Redis
		redisDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: myAppResource.Namespace,
				Name:      myAppResource.Name + "-redis-deployment",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "redis",
					},
				},
				Replicas: &myAppResource.Spec.ReplicaCount,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "redis",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "redis",
								Image: "redis:latest",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 6379,
									},
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
