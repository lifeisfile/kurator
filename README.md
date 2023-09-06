### DevOpsTrain kurator

Essential component for some courses that maintain connection with server and used for student's task validation.

If you're the course author you have to use this tool for convenient course creation. 

### Usage as learning tool

* Download pre-compiled binary for you operating system
* Login into your account using your email and password (can be acquired on https://devops.lifeisfile.com):
  * `kurator login -email <your email>`
* List available platform courses:
  * `kurator course list`
  * Notice `short name` field to use in next step
* Launch curator process to validate the result of your learning task.
  * `kurator start`
  * course specific configuration file can be passed using `-c` option with path to yaml file. Read course documentation for details
  * source dir defaults to current directory ".", so start kurator from the folder dedicated to your learning 
  * Now you may put source code into the directory and launch task validation from the web interface
  * [If you use param `--confirm-commands` please confirm the command within 2 minutes after running validate action on Devopstrain platform. - *to be implemented in next versions*]


### Usage as course development tool
* Download pre-compiled binary for you operating system
* Login into your account using your email and password (can be acquired using `kurator signup -email <your email> -name <your name>`):
  * `kurator login -email <your email>`
  * You will use this account as developer account
* Create new course yaml structure:
  * `kurator dev create-course <name>`
* Generate sample code from templates(currently only golang templates provided, not you're not limited to it):
  * `kurator dev generate-code --course_name <name> --template_path assets/templates/golang/ --output_path ../<name>-handler --module_name <golang-module-name>`
* Start dev server:
  * `kurator dev run-server --course_name <name> --handler_url http://localhost:8888/courseHandler`
* Start handler server on port 8888 


### Architecture 

#### From student standpoint

`[Students source files] <-> Kurator <-> [Devopstrain platform] -> [Course validator backend]`

#### From developer standpoint

Kurator reads course tasks yaml files and acts as web server that looks like Devopstain platform. It passes API requests to platform and your backend service for task validation and command outputs. Client(Student) must be connected to remote websocket server using `kurator start --dev` command. If any commands/file upload is defined in yaml then localweb server sends it to platform API server using /command method. This handle will work only of `--dev` is passed with `start` for client.

### Supported OS

* Linux
* MacOS (both intel and M chips)
* Windows

### Building

Just grab source code and run `go build -o kurator main.go`

### FAQ

**Q:** Can *kurator* read any personal sensitive data from my computer?  
**A:** kurator doesn't have access to files outside the specified source dir. To ensure that, you may inspect the source code and build your version of kurator. Moreover you can even launch it in your virtual enviroment like docker or any kind of virtual machine with source code shared.

**Q:** Can *kurator* launch arbitrary commands on my computer?  
**A:** No it can't. It can only run commands approved by platform. To be rest assured follow the recommendations given in previous question.