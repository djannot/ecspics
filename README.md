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

BUILD
--------------

The Dockerfile can be used to create a Docker container for this web application.

You need to modify the Dockerfile to indicate your Google Maps API key.

If you don't have a key yet, you can get one at https://developers.google.com/maps/signup

Then, you can build it and run the container.

RUN
--------------

To start the application, run ./ecspics -Namespace=<ECS Namespace> -EndPoint=<ECS endpoint using the Base URL> -Hostname=<ECS IP address>

Note that your DNS need to resolve *.<ECS Base URL>.

LICENSING
--------------

Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
