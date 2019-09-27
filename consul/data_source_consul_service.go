package consul

import (
	"fmt"
	"sort"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	catalogServiceElem = "service"

	catalogServiceCreateIndex              = "create_index"
	catalogServiceDatacenter               = "datacenter"
	catalogServiceModifyIndex              = "modify_index"
	catalogServiceNodeAddress              = "node_address"
	catalogServiceNodeID                   = "node_id"
	catalogServiceNodeMeta                 = "node_meta"
	catalogServiceNodeName                 = "node_name"
	catalogServiceServiceAddress           = "address"
	catalogServiceServiceEnableTagOverride = "enable_tag_override"
	catalogServiceServiceID                = "id"
	catalogServiceServiceName              = "name"
	catalogServiceServicePort              = "port"
	catalogServiceServiceTags              = "tags"
	catalogServiceServiceMeta              = "meta"
	catalogServiceTaggedAddresses          = "tagged_addresses"

	// Filters
	catalogServiceName = "name"
	catalogServiceTag  = "tag"
)

func dataSourceConsulService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConsulServiceRead,
		Schema: map[string]*schema.Schema{
			// Data Source Predicate(s)
			catalogServiceDatacenter: {
				// Used in the query, must be stored and force a refresh if the value
				// changes.
				Optional: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			catalogServiceTag: {
				// Used in the query, must be stored and force a refresh if the value
				// changes.
				Optional: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			catalogServiceName: {
				Required: true,
				Type:     schema.TypeString,
			},
			catalogNodesQueryOpts: schemaQueryOpts,

			// Out parameters
			catalogServiceElem: {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						catalogServiceCreateIndex: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceNodeAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceNodeID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceModifyIndex: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceNodeName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceNodeMeta: {
							Type:     schema.TypeMap,
							Computed: true,
						},
						catalogServiceServiceAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceServiceEnableTagOverride: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceServiceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceServiceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceServicePort: {
							Type:     schema.TypeString,
							Computed: true,
						},
						catalogServiceServiceTags: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						catalogServiceServiceMeta: {
							Type:     schema.TypeMap,
							Computed: true,
						},
						catalogServiceTaggedAddresses: {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									catalogNodesSchemaTaggedLAN: {
										Type:     schema.TypeString,
										Computed: true,
									},
									catalogNodesSchemaTaggedWAN: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceConsulServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := getClient(meta)

	// Parse out data source filters to populate Consul's query options
	queryOpts, err := getQueryOpts(d, client, meta)
	if err != nil {
		return errwrap.Wrapf("unable to get query options for fetching catalog services: {{err}}", err)
	}

	var serviceName string
	if v, ok := d.GetOk(catalogServiceName); ok {
		serviceName = v.(string)
	}

	var serviceTag string
	if v, ok := d.GetOk(catalogServiceTag); ok {
		serviceTag = v.(string)
	}

	// services, meta, err := client.Catalog().Services(queryOpts)
	services, meta, err := client.Catalog().Service(serviceName, serviceTag, queryOpts)
	if err != nil {
		return err
	}

	l := make([]interface{}, 0, len(services))
	for _, service := range services {
		const defaultServiceAttrs = 13
		m := make(map[string]interface{}, defaultServiceAttrs)

		m[catalogServiceCreateIndex] = fmt.Sprintf("%d", service.CreateIndex)
		m[catalogServiceModifyIndex] = fmt.Sprintf("%d", service.ModifyIndex)
		m[catalogServiceNodeAddress] = service.Address
		m[catalogServiceNodeID] = service.ID
		m[catalogServiceNodeMeta] = service.NodeMeta
		m[catalogServiceNodeName] = service.Node
		switch service.ServiceAddress {
		case "":
			m[catalogServiceServiceAddress] = service.Address
		default:
			m[catalogServiceServiceAddress] = service.ServiceAddress
		}
		m[catalogServiceServiceEnableTagOverride] = fmt.Sprintf("%t", service.ServiceEnableTagOverride)
		m[catalogServiceServiceID] = service.ServiceID
		m[catalogServiceServiceName] = service.ServiceName
		m[catalogServiceServicePort] = fmt.Sprintf("%d", service.ServicePort)
		sort.Strings(service.ServiceTags)
		m[catalogServiceServiceTags] = service.ServiceTags
		m[catalogServiceServiceMeta] = service.ServiceMeta
		m[catalogServiceTaggedAddresses] = service.TaggedAddresses

		l = append(l, m)
	}

	const idKeyFmt = "catalog-service-%s-%q-%q"
	d.SetId(fmt.Sprintf(idKeyFmt, queryOpts.Datacenter, serviceName, serviceTag))

	d.Set(catalogServiceDatacenter, queryOpts.Datacenter)
	d.Set(catalogServiceName, serviceName)
	d.Set(catalogServiceTag, serviceTag)
	if err := d.Set(catalogServiceElem, l); err != nil {
		return errwrap.Wrapf("Unable to store service: {{err}}", err)
	}

	return nil
}
