# Load Balancer vLab #
---

## 1. Brocade vTM ##

### 1.1. S3 Load Balancing ###

**Configuration tasks**

![vTM - S3 Load balancing](s3_1.jpg)


- Create a pool using the ECS node IPs. 
	- Use the right port for S3 (9020)
	- Use the right Health Monitor (HTTP)
- Create a Virtual Server, listening at port 9020, redirecting the traffic to the pool you just created
- Create a DNS Host record for your virtual server -> lb.vlab.local 
	- The ECS already has a DNS record (ecs.vlab.local)
	- You can launch the Domain Controller from mRemoteNG; it's already preconfigured.

**Verification**

- Use S3 browser to test you can access your data using both, an ECS node and the Virtual server (IP or DNS record) you just created
	- Storage Type: S3 compatible storage
	- Create an user/bucket in ECS
	- Access key ID =  ECS object user name
	- Secret Access key =  ECS secret key
	- REST endpoint = ECS node/Load Balancer:9020
	- Verity that S3 Browser is configured for HTTP
		- Browser Tools -> Options -> Connection -> Uncheck *Use Secure Transfer (HTTPS)*
	
#### 1.1.1. SSL configuration ####

![vTM - S3 SSL Load balancing](s3_2.jpg)

- [Optional] - Some applications, like CloudArray, use Virtual Style Addressing. In that case, a wildcard for the load balancer name is needed.
	- Create a wildcard **.lb.vlab.local* and verify that you can ping *bucket.lb.vlab.local*
	- You can launch the Domain Controller from mRemoteNG; it's already preconfigured.
- Modify your Virtual Server configuration to use SSL. 
	- Listening at port 9021 (HTTP) 
	- Enable SSL Decryption and create a certificate
		- Manage SSL certificates -> Create Self-Signed Certificate -> Certificate Signing Request. The Common Name in the certificate should match with the wildcard [CN = *.lb.vlab.local]

**Verification**

- Use S3 browser to test you can access your data using port 9021
	- Modify the REST endpoint configuration in your S3 Browser account = lb.vlab.local:9021
	- Verity that S3 Browser is configured for HTTPS
		- Browser Tools -> Options -> Connection -> Check *Use Secure Transfer (HTTPS)*

**Note:** When using ISVs, the certificate must be also loaded in the application, so that the communication between the application and the Load Balancer is encripted. It is recommended to offload the certificate at the Load Balancer, but it is possible to also configure end-to-end SSL encryption.




### 1.2. NFS Load Balancing ###
![vTM - NFS Load balancing](nfs_1.jpg)

- Create Pools and Virtual Servers for the TCP and UDP ports used in NFS (111, 2049 and 10000). 
	- Configure TCP Virtual Servers as *Generic Streaming* 
	- Configure UDP Virtual Servers as *UDP - Streaming*
	
At the end you should get something like this:

![vTM - NFS Load balancing](nfs_2.jpg)

- Configure sticky sessions in your NFS pools
	- Create a session persistence configuration at the IP level and associate it with your NFS pools.

**Verification**

- Use your Linux client and mount your ECS export using the Virtual Server you just created.
	- **Hint:** mount command



## 2. NGINX ##
### 2.1. S3 Load Balancing ###

![NGINX - S3 Load balancing](ngnix_1.jpg)

- Review the configuration of the NGNIX docker container, to see how local files are mapped inside the container
`cat /usr/share/oem/cloud-config.yml |grep nginxlb`


	```
		ExecStart=/bin/bash -c '/usr/bin/docker rm -f nginxlb; docker run --rm --name nginxlb -v /home/core/conf/ecs-s3-nginx.conf:/etc/nginx/nginx.conf -v /home/core/certs/server.crt:/etc/nginx/ssl/server.crt -v /home/core/certs/server.key:/etc/nginx/ssl/server.key -p 80:80 -p 443:443 nginx
	```

	As you can see, the local file */home/core/conf/ecs-s3-nginx.conf* is mapped to the NGNIX config file inside the contaner, */etc/nginx/nginx.conf*.

- Modify your local config file according to your lab configuration
	- NGNIX lb listening at port 80, redirecting traffic to ECS nodes, port 9020 (ignore SSL configuration for now).


```

more /home/core/conf/ecs-s3-nginx.conf


worker_processes 4;

events {
}

http {

  upstream ecs {
    server 10.247.7.161:9020;
  }

  server {
    listen 80;
    listen 443 default_server ssl;
    ssl_certificate         /etc/nginx/ssl/server.crt;
    ssl_certificate_key     /etc/nginx/ssl/server.key;

    location / {
      sendfile on;
      tcp_nopush on;
      tcp_nodelay on;
      client_max_body_size 1024G;
      proxy_buffering off;
      proxy_buffer_size 4k;
      proxy_pass http://ecs;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }
  }
}

```



- Restart your docker container, so that it loads the right configuration.

	`docker restart nginxlb`


**Verification**

- Use S3 browser to test that you can access your data using port 80
	- Modify the REST endpoint configuration in your S3 Browser account = lb.vlab.local:80
	- Verity that S3 Browser is configured for HTTP
		- Browser Tools -> Options -> Connection -> Uncheck *Use Secure Transfer (HTTPS)*


#### 2.1.1. SSL configuration ####

![NGNIX - S3 SSL Load balancing](ngnix_2.jpg)

- Create a certificate for the Common Name **.lb.vlab.local*
	- You can either create a new certificate using OpenSSL
		- /etc/nginx/ssl/server.crt
		- /etc/nginx/ssl/server.key
	- Or reuse the certificate you created in the previous section with Brocade.
		- In order to get the certificate created in the vTM, run
			`docker ps` -> Get the containier ID
			`docker cp <container id>:/usr/local/zeus/zxtm-11.0/conf_A/ssl/server_keys/lb_cert.public /home/core/certs/server.crt`
			`docker cp <container id>:/usr/local/zeus/zxtm-11.0/conf_A/ssl/server_keys/lb_cert.private /home/core/certs/server.key`

- Restart your docker container, so that it loads the right configuration.

	`docker restart nginxlb`


**Verification**

- Use S3 browser to test that you can access your data using port 443
	- Modify the REST endpoint configuration in your S3 Browser account = lb.vlab.local:443
	- Verity that S3 Browser is configured for HTTPS
		- Browser Tools -> Options -> Connection -> Check *Use Secure Transfer (HTTPS)*



