Auto vaas provides a progamable way to create or delete VaaS instances.   The program uses a json file in the format as follows to create or delete an instance.

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
