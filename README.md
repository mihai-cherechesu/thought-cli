### Setup
To build this command line tool, install (at least) `go1.19.4` on your machine and simply run `go build -o cpxctl`.

### Usage
This CLI currently provides only one subcommand, `ls`, that interacts with the mock cloud provider, so the usage should be pretty straightforward. Filtering can only be done with respect to
one service, by using `--service` flag. You can use `cpxctl --help` to find out more details.

### Examples
1. To get a table with all your services, one can use `cpxctl ls`. 
2. To get a table with all instances of the same service, one can use `cpxctl ls --service AuthService`. The output should be similar to:
```
┏━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━┳━━━━━━┳━━━━━━━━┳━━━━━━━━━━━┓
┃ IP          ┃ SERVICE            ┃ CPU  ┃ MEMORY ┃ STATUS    ┃
┣━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━╋━━━━━━╋━━━━━━━━╋━━━━━━━━━━━┫
┃ 10.58.1.48  ┃ RoleService        ┃ 31%  ┃ 44%    ┃ Healthy   ┃
┃ 10.58.1.72  ┃ RoleService        ┃ 90%  ┃ 72%    ┃ Healthy   ┃
┃ 10.58.1.34  ┃ RoleService        ┃ 45%  ┃ 19%    ┃ Healthy   ┃
┃ 10.58.1.124 ┃ RoleService        ┃ 95%  ┃ 70%    ┃ Unhealthy ┃
┃ 10.58.1.18  ┃ RoleService        ┃ 76%  ┃ 59%    ┃ Healthy   ┃
...
```
3. To get a merged table, one can use `cpxctl ls --merged` (the CPU and MEM are averaged and the IPs are all merged in a single list).
```
┏━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━┓
┃ IPS         ┃ SERVICE            ┃ CPU_AVG ┃ MEMORY_AVG ┃ REPLICAS ┃
┣━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━╋━━━━━━━━━━━━╋━━━━━━━━━━┫
┃ 10.58.1.48  ┃ RoleService        ┃ 39%     ┃ 42%        ┃ 25       ┃
┃ 10.58.1.72  ┃                    ┃         ┃            ┃          ┃
┃ 10.58.1.34  ┃                    ┃         ┃            ┃          ┃
┃ 10.58.1.124 ┃                    ┃         ┃            ┃          ┃
┃ 10.58.1.18  ┃                    ┃         ┃            ┃          ┃
...
```
4. To continuously monitor all instances of a service, one can use `cpxctl ls --service AuthService --follow`. Outputs should be similar to the ones above, but they are dynamically changed at a preset
interval.
5. To continously monitor all instances of a service in a merged mode, one can use `cpxctl ls --service AuthService --merged --follow`. On top of the averages that are dynamically updated, 
you can also see line charts for each of them.

![image](https://user-images.githubusercontent.com/48837715/220001447-34cb1b14-842f-4234-a9d8-010f9318fb97.png)

### Known issues
1. Using the `--merged --follow` mode and then exiting using `CTRL + C` results in the current terminal session being buggy (invisible cursor, etc). This issue has something to do with the line charts.
