# httpc

## Requirements

Please refer the [requirement document](http://aimanhanna.com/concordia/comp6461/Comp6461-F20_LA1.pdf).

The  implemented  client  should  be  named  **httpc**  (the  name  of  the  produced  executable).  
The following presents the options of your final command line.

`httpc (get|post) [-v] (-h "k:v")* [-d inline-data] [-f file] URL`

In the following, we describe the purpose of the expected httpc command options:

1.Option `-v` enables a verbose output from the command-line. Verbosity could be useful for  testing  and  debugging  stages  where  you  need  more  information  to  do  so.  You  define the format of the output. However, you are expected to print all the status, and its headers, then the contents of the response.

2.URL determines the  targeted  HTTP  server.  It  could  contain  parameters  of  the  HTTP  operation. For example, the URL 'https://www.google.ca/?q=hello+world' includes the parameter q with "hello world" value.

3.To pass the headers value to your HTTP operation, you could use -h option. The latter means setting the header of the request in the format "key: value." Notice that; you can have multiple headers by having the -h option before each header parameter.

4.-d  gives   the  user  the  possibility  to  associate  the  body  of  the  HTTP  Request  with  the  inline data, meaning a set of characters for standard input.

5.Similarly,  to -d, -f  associate  the  body  of  the  HTTP  Request  with  the  data  from  a  given  file.

6.get/post  options  are  used  to  execute  GET/POST  requests  respectively.  post  should have  either  -d  or -f  but  not  both.  However,  get  option  should  not  be  used  with  the  options -d or -f.

## Optional Tasks (Bonus Marks)

### Enhance Your HTTP Client library
In  the current  HTTP  library,  you  already  implemented  the  necessary  HTTP  specifications,  GET and POST. In this optional task, you need to implement one new specification of HTTP protocol  that  is  related  to  the  client  side.  For  example,  you  could  develop  one  of  the Redirection specifications. The latter allow your HTTP client to follow the first request with another one to new URL if the client receives a redirection code (numbers starts with 3xx). This  option  is  useful  when  the  HTTP  client  deal  with  temporally  or  parental  moved  URLs.  Notice, you are free to choose which HTTP specification to implement in your HTTP library. After selecting the specification, you should consult with the Lab Instructor before starting their implementations.

### Update The cURL Command line
Accordingly,  you  could  add  the  newly  implemented  HTTP  specifications  in  your  HTTP  library to the httpc command line. To do that, you need to create a new option that allows the user the access the newly implemented specification. In addition, you are requested to add the option –o filename, which allow the HTTP client to write the body of the response to  the  specified  file  instead  of  the  console.  For  example,  the  following  will  write  the  response to hello.txt:  

`httpc -v 'http://httpbin.org/get?course=networking&assignment=1' -o hello.txt`

## Grading Policy (10 Marks)
1. HTTP Library: total of 7 marks
    - Get request: 3 marks•Header: 1 mark 
    - Post request (eg. with body): 2 marks
    - Response (eg. parse status, code, header, and body): 1 mark
    
2. Curl-like app: total of 3 marks
    - Get command: 0.5 marks
    - Post command: 0.5 marks
    - Verbose: 0.5 marks 
    - Header: 0.5 marks
    - Inline : 0.5 marks
    - File: 0.5 marks
    
3. Optional tasks: total of 2 bonus marks
    - Supports redirect: 1.5 marks
    - Supports –o option: 0.5 marks 