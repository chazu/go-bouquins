# go-bouquins

Bouquins in Go

## TODO

* translations
* tests
* csrf
* userdb commands (init, migrate, add/remove user/email)
* error pages

## Minify

* JS: https://www.danstools.com/javascript-minify/
* CSS: curl -X POST -s --data-urlencode 'input@assets/css/bouquins.css' https://cssminifier.com/raw > assets/css/bouquins.min.css

## Deployment archive

tar czf ~/tmp/go-bouquins.tar.gz go-bouquins assets/ templates/

## Configuration

JSON config file: default ./bouquins.json, or binary argument

Example:

    {
      "calibre-path": "/usr/home/meutel/data/calibre",
      "bind-address": ":8080",
      "prod": true,
      "cookie-secret": "random",
      "external-url":"https://bouquins.meutel.net",
      "providers": [
        {
          "name": "github",
          "client-id": "ID client",
          "client-secret": "SECRET"
        },
        {
          "name": "google",
          "client-id":"ID client",
          "client-secret":"SECRET"
        }
      ]
    }

Options:

* calibre-path path to calibre data
* db-path path to calibre SQLite database (default <calibre-path>/metadata.db)
* user-db-path path to users SQLite database (default ./users.db)
* bind-address HTTP socket bind address
* prod (boolean) use minified javascript/CSS
* cookie-secret random string for cookie encryption
* external-url URL used by client browsers
* providers configuration for OAuth 2 providers
  * name provider name
  * client-id OAuth client ID
  * client-secret OAuth secret

## Users SQL

CREATE TABLE accounts (id varchar(36) PRIMARY KEY NOT NULL, name varchar(255) NOT NULL);
CREATE TABLE authentifiers (id varchar(36) NOT NULL, authentifier varchar(320) PRIMARY KEY NOT NULL, FOREIGN KEY(id) REFERENCES account(id));

