# runtil

runs some command with arguments
- when *some command with arguments* terminates, re-run *same command with same arguments*
- when there are file changes, terminate and re-run *same command with same arguments*
- when ctrl-c, terminate

## examples

repeat `sleep 3` forever

```
$ runtil /bin/sleep 3
execute: "/bin/sleep" []string{"3"}
execute: "/bin/sleep" []string{"3"}
execute: "/bin/sleep" []string{"3"}
execute: "/bin/sleep" []string{"3"}
```

run a golang program (and restart it everytime you edit the `.go` files)

```
$ runtil go run server.go
execute: "sleep" []string{"3"}
Listening to 0.0.0.0:3000
```

## issue

given a simple rails environment

```
$ rails s
=> Booting Puma
=> Rails 5.1.4 application starting in development
=> Run `rails server -h` for more startup options
Puma starting in single mode...
* Version 3.11.2 (ruby 2.4.1-p111), codename: Love Song
* Min threads: 5, max threads: 5
* Environment: development
* Listening on tcp://0.0.0.0:3000
Use Ctrl-C to stop
^C- Gracefully stopping, waiting for requests to finish
=== puma shutdown: 2018-01-26 13:50:37 +0000 ===
- Goodbye!
Exiting
```

running it under runtil causes rails to exit early

```
$ runtil rails s
execute: "rails" []string{"s"}
=> Booting Puma
=> Rails 5.1.4 application starting in development
=> Run `rails server -h` for more startup options
Exiting
execute: "rails" []string{"s"}
=> Booting Puma
=> Rails 5.1.4 application starting in development
=> Run `rails server -h` for more startup options
Exiting
```
