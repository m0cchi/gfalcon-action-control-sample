# Action Control Sample

# require
You should startup IdP on super domain.

=> https://github.com/m0cchi/gfalcon-signin-service

# run
```bash
$ export DATASOURCE='gfadmin:gfadmin@unix(/tmp/mysql.sock)/gfalcon?parseTime=true'
$ go run cmd/init.go
$ go run cmd/init_user.go
# => create user(TeamID: gfalcon, UserID: sahohime, Password: secret)
$ IDP='https://saas.m0cchi.net' PORT=50000 go run server.go
```

# License
Licensed under the MIT License.
