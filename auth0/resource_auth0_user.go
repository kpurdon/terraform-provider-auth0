package auth0

import (
	"github.com/hashicorp/terraform/helper/schema"
	auth0 "github.com/yieldr/go-auth0"
	"github.com/yieldr/go-auth0/management"
)

func newUser() *schema.Resource {
	return &schema.Resource{
		Create: createUser,
		Read:   readUser,
		Update: updateUser,
		Delete: deleteUser,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return "auth0|" + val.(string)
				},
			},
			"conn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"app_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"email_verified": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"verify_email": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"phone_verified": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func readUser(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	u, err := api.User.Read(d.Id())
	if err != nil {
		return err
	}

	for k, v := range map[string]interface{}{
		"user_id":        u.ID,
		"conn":           u.Connection,
		"username":       u.Username,
		"phone_number":   u.PhoneNumber,
		"user_metadata":  u.UserMetadata,
		"email_verified": u.EmailVerified,
		"phone_verified": u.PhoneVerified,
		"verify_email":   u.VerifyEmail,
		"app_metadata":   u.AppMetadata,
		"email":          u.Email,
		"password":       u.Password,
	} {
		if (k == "password" || k == "conn") && v.(*string) == nil {
			continue
		}

		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}

func createUser(d *schema.ResourceData, m interface{}) error {
	u := buildUser(d)
	api := m.(*management.Management)
	if err := api.User.Create(u); err != nil {
		return err
	}
	d.SetId(*u.ID)
	return nil
}

func updateUser(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	for k, v := range map[string]*management.User{
		"username":       {Username: String(d, "username"), Connection: auth0.String(d.Get("conn").(string))},
		"password":       {Password: String(d, "password"), Connection: auth0.String(d.Get("conn").(string))},
		"email":          {Email: String(d, "email"), Connection: auth0.String(d.Get("conn").(string))},
		"email_verified": {EmailVerified: Bool(d, "email_verified"), Connection: auth0.String(d.Get("conn").(string))},
	} {
		if d.HasChange(k) {
			if err := api.User.Update(d.Id(), v); err != nil {
				return err
			}
		}
	}

	u := &management.User{}

	for k, f := range map[string]func(*management.User){
		"phone_number":   func(u *management.User) { u.PhoneNumber = String(d, "phone_number") },
		"phone_verified": func(u *management.User) { u.PhoneVerified = Bool(d, "phone_verified") },
		"app_metadata":   func(u *management.User) { u.AppMetadata = Map(d, "app_metadata") },
		"user_metadata":  func(u *management.User) { u.UserMetadata = Map(d, "user_metadata") },
	} {
		if d.HasChange(k) {
			u.Connection = auth0.String(d.Get("conn").(string))

			f(u)
		}
	}

	if (*u != management.User{}) {
		if err := api.User.Update(d.Id(), u); err != nil {
			return err
		}
	}

	return readUser(d, m)
}

func deleteUser(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)

	return api.User.Delete(d.Id())
}

func buildUser(d *schema.ResourceData) *management.User {
	return &management.User{
		ID:            String(d, "user_id"),
		Connection:    String(d, "conn"),
		Username:      String(d, "username"),
		PhoneNumber:   String(d, "phone_number"),
		UserMetadata:  Map(d, "user_metadata"),
		EmailVerified: Bool(d, "email_verified"),
		VerifyEmail:   Bool(d, "verify_email"),
		PhoneVerified: Bool(d, "phone_verified"),
		AppMetadata:   Map(d, "app_metadata"),
		Email:         String(d, "email"),
		Password:      String(d, "password"),
	}
}
