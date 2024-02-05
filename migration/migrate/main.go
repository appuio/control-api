package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
)

func main() {
	ctx := context.Background()

	var dryRun, iCheckedInvitations, force bool

	flag.BoolVar(&dryRun, "dry-run", true, "dry run")
	flag.BoolVar(&iCheckedInvitations, "i-checked-invitations", false, "i checked that there are no pending invitations for the billing entities")
	flag.BoolVar(&force, "force", false, "override checks")

	flag.Parse()

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(orgv1.AddToScheme(scheme))
	utilruntime.Must(controlv1.AddToScheme(scheme))
	utilruntime.Must(billingv1.AddToScheme(scheme))
	utilruntime.Must(userv1.AddToScheme(scheme))

	c, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	old2new, new2old, err := loadMapping()
	if err != nil {
		panic(err)
	}
	_ = new2old

	var es billingv1.BillingEntityList
	if err := c.List(ctx, &es); err != nil {
		panic(err)
	}

	manifests, err := collectManifestsRequiringMigration(ctx, c)
	if err != nil {
		panic(err)
	}

	var missing []string
	for id := range manifests {
		_, ok := old2new[id]
		if ok {
			continue
		}
		missing = append(missing, id)
	}
	slices.Sort(missing)
	if len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "Missing mappings for", missing)
		if !force {
			os.Exit(1)
		}
	}

	if !iCheckedInvitations {
		fmt.Fprintln(os.Stderr, "Make sure there are no pending invitations for the billing entities")
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "All checks passed")
	fmt.Fprintln(os.Stderr, "Deleting old RBAC")
	var crbs rbacv1.ClusterRoleBindingList
	if err := c.List(ctx, &crbs); err != nil {
		panic(err)
	}
	var crs rbacv1.ClusterRoleList
	if err := c.List(ctx, &crs); err != nil {
		panic(err)
	}
	for _, crb := range crbs.Items {
		if !strings.HasPrefix(crb.Name, "billingentities-be-") {
			continue
		}
		fmt.Fprintln(os.Stderr, "Deleting binding", crb.Name)
		if err := c.Delete(ctx, &crb, client.DryRunAll); err != nil {
			panic(err)
		}
	}
	for _, cr := range crs.Items {
		if !strings.HasPrefix(cr.Name, "billingentities-be-") {
			continue
		}
		fmt.Fprintln(os.Stderr, "Deleting role", cr.Name)
		if err := c.Delete(ctx, &cr, client.DryRunAll); err != nil {
			panic(err)
		}
	}

	fmt.Fprintln(os.Stderr, "Migrating manifests")

	for id, ms := range manifests {
		if old2new[id] == "" && force {
			fmt.Fprintln(os.Stderr, "Skipping", id)
			continue
		}
		for _, m := range ms {
			switch m := m.(type) {
			case *rbacv1.ClusterRole:
				fmt.Fprintln(os.Stderr, "Skipping role", m.Name, "will be recreated by the controller")
			case *rbacv1.ClusterRoleBinding:
				opts := []client.CreateOption{}
				if dryRun {
					opts = append(opts, client.DryRunAll)
				}
				pf := roleBeRegexp.FindStringSubmatch(m.Name)
				crb := m.DeepCopy()
				crb.ObjectMeta = metav1.ObjectMeta{
					Name: "billingentities-be-" + old2new[id] + "-" + pf[2],
					Labels: map[string]string{
						"appuio.io/odoo-migrated": "true",
					},
				}
				fmt.Fprintln(os.Stderr, "Migrating role binding", m.Name, "->", crb.Name)
				if err := c.Create(ctx, crb, opts...); err != nil {
					panic(err)
				}
			case *orgv1.Organization:
				opts := []client.UpdateOption{}
				if dryRun {
					opts = append(opts, client.DryRunAll)
				}
				m.Labels["appuio.io/odoo-migrated"] = "true"
				fmt.Fprintln(os.Stderr, "Migrating org", m.Name, m.Spec.BillingEntityRef, "->", "be-"+old2new[id])
				m.Spec.BillingEntityRef = "be-" + old2new[id]
				if err := c.Update(ctx, m, opts...); err != nil {
					panic(err)
				}
			}
		}
	}
}

// loadMapping loads the mapping.csv file and compares the data with the data
func loadMapping() (old2new map[string]string, new2old map[string]string, err error) {
	old2new = make(map[string]string)
	new2old = make(map[string]string)

	cr := csv.NewReader(os.Stdin)
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read record: %w", err)
		}
		if record[1] == "" {
			fmt.Fprintln(os.Stderr, "no old id for", record[0], record[2], record[3], "found")
			continue
		}
		old2new[record[1]] = record[0]
		new2old[record[0]] = record[1]
	}
	return old2new, new2old, nil
}

var roleBeRegexp = regexp.MustCompile(`^billingentities-be-(\d+)-(.+)$`)

func collectManifestsRequiringMigration(ctx context.Context, c client.Client) (map[string][]client.Object, error) {
	manifests := map[string][]client.Object{}

	findCr := func(crs rbacv1.ClusterRoleList, name string) (rbacv1.ClusterRole, bool) {
		for _, cr := range crs.Items {
			if cr.Name == name {
				return cr, true
			}
		}
		return rbacv1.ClusterRole{}, false
	}

	var crbs rbacv1.ClusterRoleBindingList
	if err := c.List(ctx, &crbs); err != nil {
		return nil, fmt.Errorf("failed to list cluster role bindings: %w", err)
	}
	var crs rbacv1.ClusterRoleList
	if err := c.List(ctx, &crs); err != nil {
		return nil, fmt.Errorf("failed to list cluster roles: %w", err)
	}
	for _, crb := range crbs.Items {
		if strings.HasPrefix(crb.Name, "billingentities-be-") {
			if len(crb.Subjects) == 0 {
				continue
			}
			m := roleBeRegexp.FindStringSubmatch(crb.Name)
			if m == nil {
				fmt.Fprintln(os.Stderr, "can't parse", crb.Name)
				continue
			}
			id := m[1]

			manifests[id] = append(manifests[id], &crb)
			cr, ok := findCr(crs, crb.Name)
			if ok {
				manifests[id] = append(manifests[id], &cr)
			}
		}
	}

	var orgs orgv1.OrganizationList
	if err := c.List(ctx, &orgs); err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	for _, org := range orgs.Items {
		id := strings.TrimPrefix(org.Spec.BillingEntityRef, "be-")
		manifests[id] = append(manifests[id], &org)
	}

	return manifests, nil
}
