#/bin/bash

ab -p ab-post.txt -T application/json -c 6 -n 10000 http://localhost:3000/test
