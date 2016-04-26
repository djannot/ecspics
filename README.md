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

DOCKER CONTAINER
--------------

The Dockerfile can be used to create a Docker container for this web application.

A Docker container is also available in the Docker Hub (djannot/ecspics)

To run the container, you need to execute the following command:

- docker run -d -p 80:80 djannot/ecspics

Then, you need to modify the app/templates/index.tmpl to indicate your Google Maps API key.

If you don't have a key yet, you can get one at https://developers.google.com/maps/signup
