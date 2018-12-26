# ``wishlist`` - A browser based list where you can add and reserve gifts, food or whatever

Nifty utility if you're hosting an event, expect gifts or food and want to over-engineer the fuck out of it.
Prevents duplicates, enables gift suggestions and is anonymous to all users of the service.

## Usage

```bash
$ git clone https://github.com/sahib/wishlist.git
$ cd wishlist
$ make run  # should download all deps
```

## Configuration

```yaml
# version: 0 (DO NOT MODIFY THIS LINE)
auth:
  expire_time: 48h
database:
  session_cache: ./session.cache
  sqlite_path: ./data.db
mail:
  from: login@changethis.org
  smtp_host: your.smtp.host.org
  smtp_password: "my_mail_password"
  smtp_port: 465
server:
  certfile: /tmp/cert.pem
  keyfile: /tmp/key.pem
  domain: changethis.org  # The domain the certificate was issued for.
  port: 5000
```

Use ``certbot`` to easily obtain a certificate from LetsEncrypt.

## Screenshots

### Login page (``/login.html``)

<center>
<img src="https://raw.githubusercontent.com/sahib/wishlist/master/docs/list-login.png" alt="login" width="50%">
</center>

### List page (``/list.html``)

<center>
<img src="https://raw.githubusercontent.com/sahib/wishlist/master/docs/list-view.png" alt="list view" width="50%">
</center>
