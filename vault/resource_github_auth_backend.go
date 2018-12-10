package vault

import (
	"errors"
	"fmt"
	"log"
	"strings"

	// "github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/vault/api"
	// "github.com/terraform-providers/tc-terraform-provider-vault/util"
)

var githubAuthType = "github"

func githubAuthBackendResource() *schema.Resource {
	return &schema.Resource{
		Create: githubAuthBackendWrite,
		Read:   githubAuthBackendRead,
		Update: githubAuthBackendUpdate,
		Delete: githubAuthBackendDelete,

		Schema: map[string]*schema.Schema{

			"path": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "path to mount the backend",
                                Default:     githubAuthType,
				ValidateFunc: func(v interface{}, k string) (ws []string, errs []error) {
					value := v.(string)
					if strings.HasSuffix(value, "/") {
						errs = append(errs, errors.New("cannot write to a path ending in '/'"))
					}
					return
				},
			},

			"description": {
				Type:        schema.TypeString,
				Required:    false,
				ForceNew:    true,
				Optional:    true,
				Description: "The description of the auth backend",
			},

			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "The Github organization. This will be the first part of the url https://XXX.github.com.",
			},

			"base_url": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "The Github url. Examples: githubpreview.com, github.com (default)",
			},

			"ttl": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "Duration after which authentication will be expired",
			},

			"max_ttl": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "Maximum duration after which authentication will be expired",
			},
		},
	}
}

func githubAuthBackendWrite(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	authType := githubAuthType
	desc := d.Get("description").(string)
	path := d.Get("path").(string)

	log.Printf("[DEBUG] Writing auth %s to Vault", authType)

	err := client.Sys().EnableAuth(path, authType, desc)

	if err != nil {
		return fmt.Errorf("error writing to Vault: %s", err)
	}

	d.SetId(path)

	return githubAuthBackendUpdate(d, meta)
}

func githubAuthBackendDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	path := d.Id()

	log.Printf("[DEBUG] Deleting auth %s from Vault", path)

	err := client.Sys().DisableAuth(path)

	if err != nil {
		return fmt.Errorf("error disabling auth from Vault: %s", err)
	}

	return nil
}

func githubAuthBackendRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	path := d.Id()
	log.Printf("[DEBUG] Reading auth %s from Vault", path)

	present, err := isGithubAuthBackendPresent(client, path)

	if err != nil {
		return fmt.Errorf("unable to check auth backends in Vault for path %s: %s", path, err)
	}

	if !present {
		// If we fell out here then we didn't find our Auth in the list.
		d.SetId("")
		return nil
	}

	// log.Printf("[DEBUG] Reading groups for mount %s from Vault", path)
	// groups, err := githubReadAllGroups(client, path)
	// if err != nil {
	// 	return err
	// }
	// if err := d.Set("group", groups); err != nil {
	// 	return err
	// }

	// log.Printf("[DEBUG] Reading users for mount %s from Vault", path)
	// users, err := githubReadAllUsers(client, path)
	// if err != nil {
	// 	return err
	// }
	// if err := d.Set("user", users); err != nil {
	// 	return err
	// }

	return nil

}

func githubAuthBackendUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	path := d.Id()
	log.Printf("[DEBUG] Updating auth %s in Vault", path)

	configuration := map[string]interface{}{
		"base_url":        d.Get("base_url"),
		"bypass_github_mfa": d.Get("bypass_github_mfa"),
		"organization":    d.Get("organization"),
		"token":           d.Get("token"),
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		configuration["ttl"] = ttl
	}

	if maxTtl, ok := d.GetOk("max_ttl"); ok {
		configuration["max_ttl"] = maxTtl
	}

	_, err := client.Logical().Write(githubConfigEndpoint(path), configuration)
	if err != nil {
		return fmt.Errorf("error updating configuration to Vault for path %s: %s", path, err)
	}

	// if d.HasChange("group") {
	// 	oldValue, newValue := d.GetChange("group")

	// 	err = githubAuthUpdateGroups(d, client, path, oldValue, newValue)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// if d.HasChange("user") {
	// 	oldValue, newValue := d.GetChange("user")

	// 	err = githubAuthUpdateUsers(d, client, path, oldValue, newValue)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return githubAuthBackendRead(d, meta)
}

// func githubReadAllGroups(client *api.Client, path string) (*schema.Set, error) {
// 	groupNames, err := listGithubGroups(client, path)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to list groups from %s in Vault: %s", path, err)
// 	}

// 	groups := &schema.Set{F: resourceGithubGroupHash}
// 	for _, groupName := range groupNames {
// 		group, err := readGithubGroup(client, path, groupName)
// 		if err != nil {
// 			return nil, fmt.Errorf("unable to read group %s from %s in Vault: %s", path, groupName, err)
// 		}

// 		policies := &schema.Set{F: schema.HashString}
// 		for _, v := range group.Policies {
// 			policies.Add(v)
// 		}

// 		m := make(map[string]interface{})
// 		m["policies"] = policies
// 		m["group_name"] = group.Name

// 		groups.Add(m)
// 	}

// 	return groups, nil
// }

// func githubReadAllUsers(client *api.Client, path string) (*schema.Set, error) {
// 	userNames, err := listGithubUsers(client, path)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to list groups from %s in Vault: %s", path, err)
// 	}

// 	users := &schema.Set{F: resourceGithubUserHash}
// 	for _, userName := range userNames {
// 		user, err := readGithubUser(client, path, userName)
// 		if err != nil {
// 			return nil, fmt.Errorf("unable to read user %s from %s in Vault: %s", path, userName, err)
// 		}

// 		groups := &schema.Set{F: schema.HashString}
// 		for _, v := range user.Groups {
// 			groups.Add(v)
// 		}

// 		policies := &schema.Set{F: schema.HashString}
// 		for _, v := range user.Policies {
// 			policies.Add(v)
// 		}

// 		m := make(map[string]interface{})
// 		m["policies"] = policies
// 		m["groups"] = groups
// 		m["username"] = user.Username

// 		users.Add(m)
// 	}

// 	return users, nil
// }
