package clusterrolebinding

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	rbac "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestToRbacClusterRoleBindingLists(t *testing.T) {
	cases := []struct {
		ClusterRoleBindings []rbac.ClusterRoleBinding
		expected            *ClusterRoleBindingList
	}{
		{nil, &ClusterRoleBindingList{Items: []ClusterRoleBinding{}}},
		{
			[]rbac.ClusterRoleBinding{
				{
					ObjectMeta: metaV1.ObjectMeta{Name: "clusterRoleBinding"},
					Subjects: []rbac.Subject{{
						Kind:     "User",
						Name:     "dashboard",
						APIGroup: "rbac.authorization.k8s.io",
					}},
					RoleRef: rbac.RoleRef{
						APIGroup: "Role",
						Kind:     "pod-reader",
						Name:     "rbac.authorization.k8s.io",
					},
				},
			},
			&ClusterRoleBindingList{
				ListMeta: api.ListMeta{TotalItems: 1},
				Items: []ClusterRoleBinding{{
					ObjectMeta: api.ObjectMeta{Name: "clusterRoleBinding", Namespace: ""},
					TypeMeta:   api.TypeMeta{Kind: api.ResourceKindClusterRoleBinding},
				}},
			},
		},
	}
	for _, c := range cases {
		actual := toClusterRoleBindingList(c.ClusterRoleBindings, nil, dataselect.NoDataSelect)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("toRbacRoleLists(%#v) == \n%#v\nexpected \n%#v\n",
				c.ClusterRoleBindings, actual, c.expected)
		}
	}
}
