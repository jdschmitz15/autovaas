# Auto VaaS tool

## Description 

Auto vaas provides a progammable way to create or delete VaaS instances.   The program uses a json file with the following format to create or delete an instance.  

Delete only needs instance_name and delete_password.  All other variables are ignored

[
	{
		"instance_name":         "DeleteMe123",
		"owner_first_name":      "Brian",
		"owner_last_name":       "Pitta",
		"email":                 "brian.pitta@illumio.com",
		"delete_password":       "deletepassword",
		"conf_delete_password":  "deletepassword",
		"management_server":     "pce.com:855",
		"soutbound_api_version": "26",
		"unpair_existing":       "true",
		"user":                  "api_user",
		"pce_password":          "apikey",
		"conf_pce_password":     "apikey",
		"org":                   "4",
		"login_server":          "",
		"clear_existing":        "true",
	}
]

## Documentation
Run `autovaas create <json file>` to create one or more instances.

Run `autovaas delete <json file>` to delete one or more instances.