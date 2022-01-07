## Supasmuggle
A multithreaded wrapper for smuggler.py

#### Installation

`go install github.com/djIsLucid/supasmuggle@latest`

You will need to ensure that https://github.com/defparam/smuggler is installed on your system and in your path as "smuggler", otherwise this will not run.

#### Basic Usage

`supasmuggle -t 100 -f urls.txt -o urls.json"

For debugging information
`supasmuggle -t 100 -f urls.txt -o urls.json -d"

To run the exhaustive.py module (this will take much longer)
`supasmuggle -t 100 -f urls.txt -o urls.json -e"


```
Usage of supasmuggle:
  -d	Show the actual output of smuggler.py
  -e	Run exhaustive.py
  -f string
    	File containing URLs to look HRS vulnerabilities on (default "THERE IS NO SPOON")
  -o string
    	Specify an output file (default "supa_22-1-8_2022.json")
  -s int
    	Specify the time (in seconds) to wait before moving on to next host (default 360)
  -t int
    	Specify the size of the resource pool (default 50)
```


