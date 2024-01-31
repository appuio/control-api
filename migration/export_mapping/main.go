package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	odooclient "github.com/appuio/go-odoo"
)

// old code
// const roleAccountCategory = 70

// my F12ing
// const roleAccountCategory = 109

// from tobru
// const roleAccountCategory = 248

// something seems to be broken here. I only get a single record back
// var roleAccountFilter = odooclient.NewCriterion("category_id", "in", []int{roleAccountCategory})

var invoiceContactFilter = odooclient.NewCriterion("type", "=", "invoice")
var activeFilter = odooclient.NewCriterion("active", "=", true)
var notInflightFilter = odooclient.NewCriterion("vshn_control_api_inflight", "=", false)

var fetchPartnerFieldOpts = odooclient.NewOptions().FetchFields(
	"id",
	// "type",
	"name",
	"display_name",
	// "country_id",
	// "commercial_partner_id",
	// "contact_address",

	// "child_ids",
	// "user_ids",

	// "email",
	// "phone",
	// "street",
	// "street2",
	// "city",
	// "zip",
	// "country_id",

	// "parent_id",
	// "vshn_control_api_meta_status",
	"vshn_control_api_inflight",

	"x_odoo_8_ID",
)

func main() {
	session, err := odooclient.NewClient(&odooclient.ClientConfig{
		Database: "VSHNProd",
		Admin:    "odoo-automation@vshn.ch",
		Password: os.Getenv("ODOO_PASSWORD"),
		URL:      "https://central.vshn.ch/",
	})
	if err != nil {
		panic(err)
	}

	criteria := odooclient.NewCriteria().AddCriterion(activeFilter).AddCriterion(notInflightFilter).AddCriterion(invoiceContactFilter)
	accPartners, err := session.FindResPartners(criteria, fetchPartnerFieldOpts)
	if err != nil {
		panic(err)
	}

	csvw := csv.NewWriter(os.Stdout)

	for _, p := range *accPartners {
		csvw.Write([]string{
			fmt.Sprintf("%d", p.Id.Get()),
			strings.TrimPrefix(p.XOdoo8ID.Get(), "__export__.res_partner_"),
			p.Name.Get(),
			p.DisplayName.Get(),
		})
	}
	csvw.Flush()
	if err := csvw.Error(); err != nil {
		panic(err)
	}
}
