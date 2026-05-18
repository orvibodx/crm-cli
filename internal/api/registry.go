package api

import "fmt"

type EntityDef struct {
	Name      string
	Label     int
	APIPrefix string
}

var entities = map[string]EntityDef{
	"customer":    {Name: "customer", Label: 2, APIPrefix: "crmCustomer"},
	"leads":       {Name: "leads", Label: 1, APIPrefix: "crmLeads"},
	"contacts":    {Name: "contacts", Label: 3, APIPrefix: "crmContacts"},
	"business":    {Name: "business", Label: 5, APIPrefix: "crmBusiness"},
	"contract":    {Name: "contract", Label: 6, APIPrefix: "crmContract"},
	"receivables": {Name: "receivables", Label: 7, APIPrefix: "crmReceivables"},
	"plan":        {Name: "plan", Label: 8, APIPrefix: "crmReceivablesPlan"},
	"product":     {Name: "product", Label: 4, APIPrefix: "crmProduct"},
	"pool":        {Name: "pool", Label: 9, APIPrefix: "crmCustomerPool"},
	"visit":       {Name: "visit", Label: 17, APIPrefix: "crmReturnVisit"},
	"invoice":     {Name: "invoice", Label: 18, APIPrefix: "crmInvoice"},
}

func GetEntity(name string) (EntityDef, error) {
	e, ok := entities[name]
	if !ok {
		return EntityDef{}, fmt.Errorf("unknown entity: %s (valid: customer, leads, contacts, business, contract, receivables, plan, product, pool, visit, invoice)", name)
	}
	return e, nil
}

func EntityNames() []string {
	names := make([]string, 0, len(entities))
	for k := range entities {
		names = append(names, k)
	}
	return names
}
