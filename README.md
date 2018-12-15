# HashiCorp Vault plugin for Google Auth.

A HashiCorp Vault plugin for Google Auth.

## Setup

The setup guide assumes some familiarity with Vault and Vault's plugin
ecosystem. You must have a Vault server already running, unsealed, and
authenticated.

1. Compile the plugin from source.

* When building remember your target platform.

  e.g. on MacOS targeting Linux:
  ```sh
  GOOS=linux make
  ```

2. Move the compiled plugin into Vault's configured `plugin_directory`:

   ```sh
   $ mv google-auth-vault-plugin /etc/vault.d/plugins/google-auth-vault-plugin
   ```

* Ensure the plugin has the IPC_CAP (shared memory) as well as vault.

  e.g.
  ```sh
  $ sudo setcap cap_ipc_lock=+ep /etc/vault.d/plugins/google-auth-vault-plugin
  ```

* You need to set [api_addr](https://www.vaultproject.io/docs/configuration/index.html#api_addr)

  This can be set at the top level for a standalone setup, or in a ha_storage stanza.

```json
api_addr          = "https://vault.mydomain.net:8200"
```

1. Calculate the SHA256 of the plugin and register it in Vault's plugin catalog.
If you are downloading the pre-compiled binary, it is highly recommended that
you use the published checksums to verify integrity.

   ```sh
   $ export SHA256=$(shasum -a 256 "/etc/vault.d/plugins/google-auth-vault-plugin" | cut -d' ' -f1)
   $ vault write sys/plugins/catalog/google-auth-vault-plugin \
       sha_256="${SHA256}" \
       command="google-auth-vault-plugin"
   ```

1. Mount the auth method:

   ```sh
   $ vault auth enable \
       -path="google" \
       -plugin-name="google-auth-vault-plugin" plugin
   ```

1. Create an OAuth client ID in [the Google Cloud Console](https://console.cloud.google.com/apis/credentials), of type "Other".

1. Configure the auth method:

   ```sh
   $ vault write auth/google/config \
       client_id=<GOOGLE_CLIENT_ID> \
       client_secret=<GOOGLE_CLIENT_SECRET>
   ```

1. Create a role for a given set of Google users mapping to a set of policies:

   Create a policy called hello: [vault polices](https://www.vaultproject.io/intro/getting-started/policies.html)

   ```sh
   $ vault write auth/google/role/hello \
       bound_domain=<DOMAIN> \
       bound_emails=myuseremail@<DOMAIN>,otheremail@<DOMAIN> \
       policies=hello
   ```

   The plugin can also map users to policies via Google Groups, however this needs a bit more setup (more info below).

   Alternative auth method with groups enabled:
   ```sh
   $ vault write auth/google/config \
       client_id=<GOOGLE_CLIENT_ID> \
       client_secret=<GOOGLE_CLIENT_SECRET> \
       fetch_groups=true
       impersonation=some.admin@your-google-domain.com
       admin_service_account=base64-encoded-service-account-json-file
   ```

   Create a role for a Google group mapping to a set of policies:
   ```sh
   $ vault write auth/google/role/hello \
       bound_domain=<DOMAIN> \
       bound_groups=SecurityTeam,WebTeam \
       policies=hello
   ```

1. Login using Google credentials (NB we use `open` to navigate to the Google Auth URL to get the code).

   ```sh
   $ open $(vault read -field=url auth/google/code_url)
   $ vault write auth/google/login code=$GOOGLE_CODE role=hello
   ```


## Setup of group retrieval

To get groups information the plugin needs to access the Google Admin Directory API which is only possible for users with admin privileges. The plugin cannot use the user provided token to do this because this would mean that all authenticating users would need permission to the admin API. Therefore a serivce account with G Suite Domain-Wide Delegation of Authority is needed.

Please follow this guide on how to create one and authorize it: https://developers.google.com/admin-sdk/directory/v1/guides/delegation

You don't have to grant the read/write scopes that the guide tells you to, this plugin only needs `https://www.googleapis.com/auth/admin.directory.group.readonly`.
Once this is done you can fill in the config values from the `groups_enabled` example above.

The base64 `admin_service_account` value can be generated like this:

```sh
cat your-json-service-account-file | base64 -w0`.
```

The `impersonation` field needs to be the Google account email address of a real admin user which the service account will impersonate.

Explanation from the Google API Docs:

> Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore your service account needs to impersonate one of those users to access the Admin SDK Directory API. Additionally, the user must have logged in at least once and accepted the G Suite Terms of Service.


## License

This code is licensed under the MPLv2 license.
