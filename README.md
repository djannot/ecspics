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

DOCKER CONTAINER
--------------

The Dockerfile can be used to create a Docker container for this web application.

You need to modify the Dockerfile to indicate your Google Maps API key.

If you don't have a key yet, you can get one at https://developers.google.com/maps/signup

Then, you can build it and run the container.

To start the application, run ./ecspics -Namespace=<ECS Namespace> -EndPoint=<ECS endpoint using the Base URL> -Hostname=<ECS IP address>

Note that your DNS need to resolve *.<ECS Base URL>.
