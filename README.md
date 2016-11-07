ECSPics
==============

[![Build Status](https://drone.io/github.com/djannot/ecspics/status.png)](https://drone.io/github.com/djannot/ecspics/latest)

[![wercker status](https://app.wercker.com/status/ee4a5af9191817fee967b3ad43ddaf84/m "wercker status")](https://app.wercker.com/project/bykey/ee4a5af9191817fee967b3ad43ddaf84)

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
docker run -p 8080:80 -e PORT=80 ecspics

The application will be available on http://\<ip of application host\>:8080

LICENSING
--------------

Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
