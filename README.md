ECSPics
==============

OVERVIEW
--------------

ECSPics is a web application developped in Golang and leveraging AngularJS

It's a way to share best practices:

- Expose the features through a REST API
- Keep the web server outside of the data path

And also a way to show some ECS unique capabilities:

- Active Directory self service
- Metadata search
- The ability to apply retentions to object

You need to create a Base URL with namespace on ECS because the application is using CORS.
Note that your DNS need to resolve *.\<ECS Base URL\>.

BUILD
--------------

The Dockerfile can be used to create a Docker container for this web application.

Just run the following command in the folder that contains the Dockerfile: docker build -t ecspics . 

Please note that you have to correct the Namespace, EndPoint and Hostname either in the Dockerfile or when running the container.

RUN
--------------

To start the application, run: 
docker run -p 8080:80 -e "PORT=80 -e HOSTNAME=10.10.10.1 -e ENDPOINT=http://ecs-vip.local.net:9020 -e NAMESPACE=ns01 ecspics

The parameters are the ECS Hostname or IP, the ECS Endpoint (or Loadbalancer) and the Namespace to use.

The application will be available on http://\<ip of application host\>:8080

LICENSING
--------------

Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
