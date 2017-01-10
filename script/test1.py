#/usr/bin/env python
#coding=utf8

import socket
import time
import uuid

self_jid = 'test1@a.a'
to_jid = 'test2@a.a'

p = '''{"envelope":{"id":"1234567890","type":0,"jid":"%s","pwd":"abc123"},"vsn":"0.0.1"}\01'''
p1 ='''
{
"envelope": { "id": "%s", "type": 1, "from": "%s", "to": "%s" },
"payload": { "content": "hi girl" }
}
\01'''
print p

client = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
client.connect(("127.0.0.1",4222))
client.send(p % self_jid)

try:
    print "recv:",client.recv(1024)
    for i in range(0,1):
        #print "recv:",client.recv(1024)
        packet = p1 % (uuid.uuid4().hex,self_jid,to_jid)
        print i,"👮 send:",packet
        client.send(packet)
        print "----------------------"
        time.sleep(1)
except :
    print "close"
print "bye bye"
client.close()
