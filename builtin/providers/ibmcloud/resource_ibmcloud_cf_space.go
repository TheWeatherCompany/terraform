package ibmcloud

import (
	"fmt"

	"github.com/IBM-Bluemix/bluemix-go/api/cf/cfv2"
	"github.com/IBM-Bluemix/bluemix-go/bmxerror"
	"github.com/IBM-Bluemix/bluemix-go/helpers"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceIBMCloudCfSpace() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMCloudCfSpaceCreate,
		Read:     resourceIBMCloudCfSpaceRead,
		Update:   resourceIBMCloudCfSpaceUpdate,
		Delete:   resourceIBMCloudCfSpaceDelete,
		Exists:   resourceIBMCloudCfSpaceExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name for the space",
			},
			"org": {
				Description: "The org this space belongs to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"space_quota": {
				Description: "The name of the Space Quota Definition",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceIBMCloudCfSpaceCreate(d *schema.ResourceData, meta interface{}) error {
	spaceClient, err := meta.(ClientSession).CloudFoundrySpaceClient()
	if err != nil {
		return err
	}
	orgClient, _ := meta.(ClientSession).CloudFoundryOrgClient()
	org := d.Get("org").(string)
	name := d.Get("name").(string)

	req := cfv2.SpaceCreateRequest{
		Name: name,
	}

	orgFields, err := orgClient.FindByName(org)
	if err != nil {
		return fmt.Errorf("Error retrieving org: %s", err)
	}
	req.OrgGUID = orgFields.GUID

	if spaceQuota, ok := d.GetOk("space_quota"); ok {
		spaceQuotaClient, _ := meta.(ClientSession).CloudFoundrySpaceQuotaClient()
		quota, err := spaceQuotaClient.FindByName(spaceQuota.(string), orgFields.GUID)
		if err != nil {
			return fmt.Errorf("Error retrieving space quota: %s", err)
		}
		req.SpaceQuotaGUID = quota.GUID
	}

	space, err := spaceClient.Create(req)
	if err != nil {
		return fmt.Errorf("Error creating space: %s", err)
	}

	d.SetId(space.Metadata.GUID)
	return resourceIBMCloudCfSpaceRead(d, meta)
}

func resourceIBMCloudCfSpaceRead(d *schema.ResourceData, meta interface{}) error {
	spaceClient, err := meta.(ClientSession).CloudFoundrySpaceClient()
	if err != nil {
		return err
	}
	spaceGUID := d.Id()

	_, err = spaceClient.Get(spaceGUID)
	if err != nil {
		return fmt.Errorf("Error retrieving space: %s", err)
	}
	return nil
}

func resourceIBMCloudCfSpaceUpdate(d *schema.ResourceData, meta interface{}) error {
	spaceClient, err := meta.(ClientSession).CloudFoundrySpaceClient()
	if err != nil {
		return err
	}
	id := d.Id()

	req := cfv2.SpaceUpdateRequest{}
	if d.HasChange("name") {
		req.Name = helpers.String(d.Get("name").(string))
	}

	_, err = spaceClient.Update(id, req)
	if err != nil {
		return fmt.Errorf("Error updating space: %s", err)
	}

	return resourceIBMCloudCfSpaceRead(d, meta)
}

func resourceIBMCloudCfSpaceDelete(d *schema.ResourceData, meta interface{}) error {
	spaceClient, err := meta.(ClientSession).CloudFoundrySpaceClient()
	if err != nil {
		return err
	}
	id := d.Id()

	err = spaceClient.Delete(id)
	if err != nil {
		return fmt.Errorf("Error deleting space: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceIBMCloudCfSpaceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	spaceClient, err := meta.(ClientSession).CloudFoundrySpaceClient()
	if err != nil {
		return false, err
	}
	id := d.Id()

	space, err := spaceClient.Get(id)
	if err != nil {
		if apiErr, ok := err.(bmxerror.RequestFailure); ok {
			if apiErr.StatusCode() == 404 {
				return false, nil
			}
		}
		return false, fmt.Errorf("Error communicating with the API: %s", err)
	}

	return space.Metadata.GUID == id, nil
}
