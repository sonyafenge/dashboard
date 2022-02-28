package iam

import (
	"errors"
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/args"
	"github.com/kubernetes/dashboard/src/app/backend/client"
	"github.com/kubernetes/dashboard/src/app/backend/iam/db"
	"github.com/kubernetes/dashboard/src/app/backend/iam/model"
	"github.com/kubernetes/dashboard/src/app/backend/resource/clusterrolebinding"
	ns "github.com/kubernetes/dashboard/src/app/backend/resource/namespace"
	"github.com/kubernetes/dashboard/src/app/backend/resource/serviceaccount"
	rbac "k8s.io/api/rbac/v1"
	"log"
	"os"
	"strings"
	"time"
)

// Create cluster Admin

func CreateClusterAdmin() error {
	const adminName = "centaurus"
	const dashboardNS = "centaurus-dashboard"
	const clusterRoleName = "cluster-admin"
	const saName = adminName + "-dashboard-sa"
	admin := os.Getenv("CLUSTER_ADMIN")
	if admin == "" {
		admin = adminName
	}
	clientManager := client.NewClientManager(args.Holder.GetKubeConfigFile(), args.Holder.GetApiServerHost())

	// TODO Check if centaurus-dashboard namespace exists or not using GET method
	k8sClient := clientManager.InsecureClient()

	// Create namespace
	namespaceSpec := new(ns.NamespaceSpec)
	namespaceSpec.Name = dashboardNS
	if err := ns.CreateNamespace(namespaceSpec, "system", k8sClient); err != nil {
		log.Printf("Create namespace for admin user failed, err:%s ", err.Error())
		//return err
	} else {
		log.Printf("Create Namespace successfully")
	}

	// Create Service Account
	serviceAccountSpec := new(serviceaccount.ServiceAccountSpec)
	serviceAccountSpec.Name = saName
	serviceAccountSpec.Namespace = dashboardNS
	if err := serviceaccount.CreateServiceAccount(serviceAccountSpec, k8sClient); err != nil {
		log.Printf("Create service account for admin user failed, err:%s ", err.Error())
		//return err
	}

	// Create Cluster Role Binding
	clusterRoleBindingSpec := &clusterrolebinding.ClusterRoleBindingSpec{
		Name: "admin-cluster-role-binding",
		Subject: rbac.Subject{
			Kind:      "ServiceAccount",
			APIGroup:  "",
			Name:      saName,
			Namespace: dashboardNS,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
	}
	if err := clusterrolebinding.CreateClusterRoleBindings(clusterRoleBindingSpec, k8sClient); err != nil {
		log.Printf("Create cluster role binding for admin user failed, err:%s ", err.Error())
		// return err
	}

	// Get Token
	var found = false
	var retrial = 0
	var token []byte
	for {
		if retrial == 4 && !found {
			break
		}
		secretList, err := k8sClient.CoreV1().SecretsWithMultiTenancy(dashboardNS, "").List(api.ListEverything)
		if err != nil {
			log.Printf("Get secret for admin user failed, err:%s \n", err.Error())
			return errors.New("secret not found for admin user")
		}

		for _, secret := range secretList.Items {
			checkName := strings.Contains(secret.Name, saName)
			if secret.Namespace == dashboardNS && checkName {
				token = secret.Data["token"]
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(1)
	}
	if !found && retrial == 4 {
		log.Printf("Get token for admin user failed after 3 retrial, err:%s \n", "Get token failed")
		return nil
	}

	// Create User and enter data into DB
	user := model.User{
		ID:                0,
		Username:          admin,
		Password:          "Centaurus@123",
		Token:             string(token),
		Type:              "cluster-admin",
		Tenant:            "system",
		Role:              "",
		NameSpace:         "default",
		CreationTimestamp: time.Now(),
	}

	// call insertUser function and pass the user data
	insertID := db.InsertUser(user)

	log.Printf("\nUser Id: %d", insertID)
	return nil
}
