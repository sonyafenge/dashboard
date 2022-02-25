package role

import (
	"reflect"
	"testing"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	rbac "k8s.io/api/rbac/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestToRbacRoleLists(t *testing.T) {
	cases := []struct {
		Roles    []rbac.Role
		expected *RoleList
	}{
		{nil, &RoleList{Items: []Role{}}},
		{
			[]rbac.Role{
				{
					ObjectMeta: metaV1.ObjectMeta{Name: "role"},
					Rules: []rbac.PolicyRule{{
						Verbs:     []string{"post", "put"},
						Resources: []string{"pods", "deployments"},
					}},
				},
			},
			&RoleList{
				ListMeta: api.ListMeta{TotalItems: 1},
				Items: []Role{{
					ObjectMeta: api.ObjectMeta{Name: "role", Namespace: ""},
					TypeMeta:   api.TypeMeta{Kind: api.ResourceKindRole},
				}},
			},
		},
	}
	for _, c := range cases {
		actual := toRoleList(c.Roles, nil, dataselect.NoDataSelect)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("toRbacRoleLists(%#v) == \n%#v\nexpected \n%#v\n",
				c.Roles, actual, c.expected)
		}
	}
}
