# stv-web
Single Transferable Vote web application

An email sending, AD authed very much over-engineered STV system that Liam made because he could!

All public and template files are included in the executable/container so no need for copying files over.

## How to build

Use the command `make` in order to make the executable file for your system.

If you need to change the OS or Architecture of the executable for another system then use e.g. `GOOS=linux GOARCH=amd64 make`, change the GOOS and GOARCH as needed for your desired system.

For docker there is a docker file so use this command to make the container `docker run -p 6691:6691 --name ystv-stv-web -v <location fo db>:/db -v <location of toml>:/toml --restart=always ystv-stv-web:latest`

The DB folder can be left empty as there will be a db file created, for the TOML folder then use the example config.toml for reference.