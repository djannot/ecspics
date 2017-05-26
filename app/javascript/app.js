var thumbnail_canvas;
var image_canvas;
var image_data;

// Extract EXIF metadatas from the picture
var extractMetadata = function(exifObject) {
  console.log(exifObject);
  if(exifObject["ImageWidth"]) {
    $("#image_width").val(exifObject["ImageWidth"]);
  }
  if(exifObject["ImageHeight"]) {
    $("#image_height").val(exifObject["ImageHeight"]);
  }
  latitude = parseFloat(exifObject["GPSLatitude"][0]) + parseFloat(exifObject["GPSLatitude"][1]) / 60 + parseFloat(exifObject["GPSLatitude"][2]) / 3600;
  if(exifObject["GPSLatitudeRef"] == "S") {
    latitude = - latitude;
  }
  longitude = parseFloat(exifObject["GPSLongitude"][0]) + parseFloat(exifObject["GPSLongitude"][1]) / 60 + parseFloat(exifObject["GPSLongitude"][2]) / 3600;
  if(exifObject["GPSLongitudeRef"] == "W") {
    longitude = - longitude;
  }
  $("#gps_latitude").val(latitude);
  $("#gps_longitude").val(longitude);
  $("#datetime").val(exifObject["DateTime"]);
}

// Clone a canvas with a different scale
function getCanvas(original, scale) {
  var canvas = document.createElement("canvas");
  canvas.width = original.width * scale;
  canvas.height = original.height * scale;
  canvas.getContext("2d").drawImage(original, 0, 0, canvas.width, canvas.height);
  return canvas
}

if (!String.prototype.encodeHTML) {
  String.prototype.encodeHTML = function () {
    return this.replace(/&/g, '&amp;')
               .replace(/</g, '&lt;')
               .replace(/>/g, '&gt;')
               .replace(/"/g, '&quot;')
               .replace(/'/g, '&apos;');
  };
}

(function() {
  var app = angular.module('ECSPics', ['ngAnimate', 'ngSanitize', 'angularUtils.directives.dirPagination']);

  app.value('loadingService', {
    loadingCount: 0,
    isLoading: function() { return loadingCount > 0; },
    requested: function() { loadingCount += 1; },
    responded: function() { loadingCount -= 1; }
  });

  app.factory('loadingInterceptor', ['$q', 'loadingService', function($q, loadingService) {
    return {
      request: function(config) {
        loadingService.requested();
        return config;
      },
      response: function(response) {
        loadingService.responded();
        return response;
      },
      responseError: function(rejection) {
        loadingService.responded();
        return $q.reject(rejection);
      },
    }
  }]);

  app.config(["$httpProvider", function ($httpProvider) {
    $httpProvider.interceptors.push('loadingInterceptor');
  }]);

  // Main controller that calls APIs to get the list of buckets and information about ECS
  app.controller('PicsController', ['$http', '$animate', '$scope', 'loadingService', 'picsService', function($http, $animate, $scope, loadingService, picsService) {
    $scope.pics = picsService;
    loadingCount = 0;
    $scope.loadingService = loadingService;
    $scope.pics.buckets = [];
    $scope.pics.hostname = "";
    $scope.information = 0;
    $http.get('/api/v1/buckets').success(function(data) {
      $scope.pics.buckets = data;
    }).
    error(function(data, status, headers, config) {
      $scope.pics.messagetitle = "Error";
      $scope.pics.messagebody = data;
      $('#message').modal('show');
    });
  }]);

  app.factory('picsService', function() {
    return {}
  });

  // Upload a picture and a thumbnail
  app.directive("picsUpload", function() {
    return {
      restrict: 'E',
      templateUrl: "app/html/pics-upload.html",
      controller: ['$http', '$scope', 'picsService', function($http, $scope, picsService) {
        $scope.pics = picsService;
        $scope.pics.image  = new Image();
        // Send information about the picture to let the server compute the signatures
        this.uploadPicture = function(pics) {
          $http.post('/api/v1/uploadpicture', {
            bucket: $("#bucket").val(),
            retention: $("#retention").val(),
            file_size: $("#file_size").val(),
            file_name: $("#file_name").val(),
            image_width: $("#image_width").val(),
            image_height: $("#image_height").val(),
            gps_latitude: $("#gps_latitude").val(),
            gps_longitude: $("#gps_longitude").val(),
            datetime: $("#datetime").val()
          }).
            success(function(data, status, headers, config) {
              $('#upload_thumbnail_item > span > i').removeClass().addClass("fa fa-refresh fa-spin");
              $('#upload_thumbnail_item').show();
              $('#upload_picture_item > span > i').removeClass().addClass("fa fa-refresh fa-spin");
              $('#upload_picture_item').show();
              $scope.pics.messagetitle = "Upload in progress";
              $scope.pics.messagebody = '<h3>JSON data received from the server to upload the file from the web browser (including the signature computed by the server):</h3><pre><code>' + JSON.stringify(data, undefined, 2) + '</code></pre>';
              $('#message').modal({show: true});
              //setTimeout(function() { $('#message').modal('hide'); }, 10000);
              $scope.uploadCtrl.executeUpload(data);
            }).
            error(function(data, status, headers, config) {
              $('#upload_thumbnail_item > span > i').removeClass().addClass("glyphicon glyphicon-remove");
              $('#upload_thumbnail_item').show();
              $('#upload_picture_item > span > i').removeClass().addClass("glyphicon glyphicon-remove");
              $('#upload_picture_item').show();
              $scope.pics.messagetitle = "Error";
              $scope.pics.messagebody = "Can't get javascript code from the server to upload the file";
              $('#message').modal({show: true});
            });
        };
        // Upload the picture and thumbnail using the signatures provided by the server
        this.executeUpload = function(data) {
          var files = $("#file")[0].files;
          var pictureReader = new FileReader();
          pictureReader.onload = function(event) {
            var content = event.target.result;
            try {
              var pictureHeaders = {};
              for (var header in data["picture_headers"]) {
                pictureHeaders[header] = data["picture_headers"][header][0];
              }
              $http({
                url: data["picture_url"],
                method: 'PUT',
                headers: pictureHeaders,
                data: new Uint8Array(content),
                transformRequest: []
              }).
                success(function(data, status, headers, config) {
                  $('#upload_picture_item > span > i').removeClass().addClass("glyphicon glyphicon-ok");
                  $('#message').modal('hide');
                }).
                error(function(data, status, headers, config) {
                  $('#upload_picture_item > span > i').removeClass().addClass("glyphicon glyphicon-remove");
                  $scope.pics.messagetitle = "Error";
                  $scope.pics.messagebody = "Picture upload failed";
                  $('#message').modal({show: true});
                });
            }
            catch (e) {
              $('#upload_picture_item > span > i').removeClass().addClass("glyphicon glyphicon-remove");
              $scope.pics.messagetitle = "Error";
              $scope.pics.messagebody = "Picture upload failed";
              $('#message').modal({show: true});
            }
          }
          if(files[0]) {
            pictureReader.readAsArrayBuffer(files[0]);
          } else {
            pictureReader.readAsArrayBuffer(image_data);
          }


          thumbnail_canvas.toBlob(function(blob) {
            var thumbnailReader = new FileReader();
            thumbnailReader.onload = function(event) {
              var content = event.target.result;
              try {
                var thumbnailHeaders = {};
                for (var header in data["thumbnail_headers"]) {
                  thumbnailHeaders[header] = data["thumbnail_headers"][header][0];
                }
                $http({
                  url: data["thumbnail_url"],
                  method: 'PUT',
                  headers: thumbnailHeaders,
                  data: new Uint8Array(content),
                  transformRequest: []
                }).
                  success(function(data, status, headers, config) {
                    $('#upload_thumbnail_item > span > i').removeClass().addClass("glyphicon glyphicon-ok");
                  }).
                  error(function(data, status, headers, config) {
                    $('#upload_thumbnail_item > span > i').removeClass().addClass("glyphicon glyphicon-remove");
                    $scope.pics.messagetitle = "Error";
                    $scope.pics.messagebody = "Thumbnail upload failed";
                    $('#message').modal({show: true});
                  });
              }
              catch (e) {
                $('#upload_thumbnail_item > span > i').removeClass().addClass("glyphicon glyphicon-remove");
                $scope.pics.messagetitle = "Error";
                $scope.pics.messagebody = "Thumbnail upload failed";
                $('#message').modal({show: true});
              }
            }
            thumbnailReader.readAsArrayBuffer(blob);
          });
        };
        // Extract information from the picture
        this.getInformation = function() {
          $('#extract_metadata_item').hide();
          $('#create_thumbnail_item').hide();
          $('#upload_thumbnail_item').hide();
          $('#upload_picture_item').hide();
          $('#create_thumbnail_item > span > i').removeClass().addClass("fa fa-refresh fa-spin");
          $('#create_thumbnail_item').show();
          $("#gps_latitude").val("");
          $("#gps_longitude").val("");
          $("#datetime").val("");
          $("#file_name").val($("#picture_url").val().split('?')[0].substring($("#picture_url").val().lastIndexOf('/')+1));
          $http.get($("#picture_url").val(), {responseType: 'blob'}).
            success(function(data, status, headers, config) {
              $("#file_size").val(data.size);
              image_data = data;
              $scope.pics.image.crossOrigin = "anonymous";
              $scope.pics.image.onload = function() {
                $("#image_width").val($scope.pics.image.width);
                console.log($scope.pics.image.width);
                $("#image_height").val($scope.pics.image.height);
                thumbnail_canvas = getCanvas($scope.pics.image, 1/10);
                $('#create_thumbnail_item > span > i').removeClass().addClass("glyphicon glyphicon-ok");
                EXIF.getData($scope.pics.image, function() {
                  extractMetadata(EXIF.getAllTags($scope.pics.image));
                });
                $scope.information = 1;
              };
              $scope.pics.image.src = window.URL.createObjectURL(new Blob([data]));
            }).
            error(function(data, status, headers, config) {
              $scope.pics.messagetitle = "Error";
              $scope.pics.messagebody = data;
              $('#message').modal({show: true});
            });
          $('#extract_metadata_item > span > i').removeClass().addClass("fa fa-refresh fa-spin");
          $('#extract_metadata_item').show();

          $('#extract_metadata_item > span > i').removeClass().addClass("glyphicon glyphicon-ok");
        };
      }],
      controllerAs: "uploadCtrl"
    };
  });

  // Crate a new bucket
  app.directive("picsBucket", function() {
    return {
      restrict: 'E',
      templateUrl: "app/html/pics-bucket.html",
      controller: ['$http', '$scope', 'picsService', function($http, $scope, picsService) {
        $scope.pics = picsService;
        this.createBucket = function() {
          $http.post('/api/v1/createbucket', {bucket: this.bucket, encrypted: false}).
            success(function(data, status, headers, config) {
              $scope.pics.buckets.push(data["bucket"]);
              $scope.pics.messagetitle = "Success";
              $scope.pics.messagebody = "Bucket created with the following CORS configuration:<br /><br /><pre class='prettyprint'><code class='language-xml'>" + data["cors_configuration"].encodeHTML() + "</pre></code>";
              $('#message').modal({show: true});
            }).
            error(function(data, status, headers, config) {
              $scope.pics.messagetitle = "Error";
              $scope.pics.messagebody = data;
              $('#message').modal({show: true});
            });
        };
      }],
      controllerAs: "bucketCtrl"
    };
  });

  // Send Metadata Search request
  app.directive("picsSearch", function() {
    return {
      restrict: 'E',
      templateUrl: "app/html/pics-search.html",
      controller: ['$http', '$scope', 'picsService', function($http, $scope, picsService) {
        $scope.pics = picsService;
        $scope.currentPage = 1;
        $scope.pics.pictures = {};
        $scope.pics.markers = [];
        this.searchPics = function(pics) {
          for (var i = 0; i < $scope.pics.markers.length; i++) {
            $scope.pics.markers[i].setMap(null);
          }
          $scope.pics.markers = [];
          var sw = $scope.pics.rectangle.getBounds().getSouthWest();
          var ne = $scope.pics.rectangle.getBounds().getNorthEast();
          if(sw.lng() > ne.lng()) {
            $scope.pics.pictures = [];
            $scope.pics.messagetitle = "Error";
            $scope.pics.messagebody = "Can't select an area that is crossing the globe";
            $('#message').modal({show: true});
          } else {
            $http.post('/api/v1/search', {
              search_bucket: $("#search_bucket").val(),
              search_width: this.search_width,
              search_height: this.search_height,
              search_area: this.search_area,
              search_sw_latitude: sw.lat().toString(),
              search_sw_longitude: sw.lng().toString(),
              search_ne_latitude: ne.lat().toString(),
              search_ne_longitude: ne.lng().toString()
            }).
              success(function(data, status, headers, config) {
                $scope.pics.pictures = data;
                for (var i=0; i < $scope.pics.pictures.length; i++) {
                  var index = i;
                  if (($scope.pics.pictures[i]["picture_metadatas"]["x-amz-meta-gps-latitude"] != "") && ($scope.pics.pictures[i]["picture_metadatas"]["x-amz-meta-gps-longitude"] != "")) {
                    var myLatLng = {lat: parseFloat($scope.pics.pictures[i]["picture_metadatas"]["x-amz-meta-gps-latitude"]), lng: parseFloat($scope.pics.pictures[i]["picture_metadatas"]["x-amz-meta-gps-longitude"])};
                    var marker = new google.maps.Marker({
                      position: myLatLng,
                      map: $scope.pics.bigmap,
                      title: index.toString()
                    });
                    $scope.pics.markers.push(marker);
                    marker.addListener('click', function() {
                      title = this.title;
                      $scope.$apply(function() {
                        $scope.showCtrl.displayPicture(parseInt(title));
                      });
                    });
                  }
                }
              }).
              error(function(data, status, headers, config) {
                $scope.pics.pictures = [];
                $scope.pics.messagetitle = "Error";
                $scope.pics.messagebody = data;
                $('#message').modal({show: true});
              });

          }
        };
      }],
      controllerAs: "searchCtrl"
    };
  });

  // Display pictures
  app.directive("picsShow", function() {
    return {
      restrict: 'E',
      templateUrl: "app/html/pics-show.html",
      controller: ['$http', '$scope', 'picsService', function($http, $scope, picsService) {
        $scope.pics = picsService;
        this.displayPicture = function(index) {
          $scope.pics.messagetitle = $scope.pics.pictures[index]["picture_key"];
          $scope.pics.messagebody = '<div class="picture"><img src="' + $scope.pics.pictures[index]["picture_url"] + '" /><br /><ul class="list-group"><li class="list-group-item">Url: <a href="' + $scope.pics.pictures[index]["picture_url"] + '">' + $scope.pics.pictures[index]["picture_url"] + '</a></li>';
          if(typeof $scope.pics.pictures[index]["picture_metadatas"]["x-amz-meta-gps-latitude"] != "undefined") {
            $scope.pics.messagebody += '<li class="list-group-item">Latitude: ' + $scope.pics.pictures[index]["picture_metadatas"]["x-amz-meta-gps-latitude"] + '</li>';
          }
          if(typeof $scope.pics.pictures[index]["picture_metadatas"]["x-amz-meta-gps-longitude"] != "undefined") {
            $scope.pics.messagebody += '<li class="list-group-item">Longitude: ' + $scope.pics.pictures[index]["picture_metadatas"]["x-amz-meta-gps-longitude"] + '</li>';
          }
          $scope.pics.messagebody += '</ul></div>';
          $('#message').modal({show: true});
        };
        this.deletePicture = function(index) {
          pictureHeaders = {};
          for (var header in $scope.pics.pictures[index]["delete_request_picture_headers"]) {
            pictureHeaders[header] = $scope.pics.pictures[index]["delete_request_picture_headers"][header][0];
          }
          $http({
            url: $scope.pics.pictures[index]["delete_request_picture_url"],
            method: "DELETE",
            headers: pictureHeaders
          }).
            success(function(data, status, headers, config) {
              $scope.pics.pictures.splice(index, 1);
              for (var i = 0; i < $scope.pics.markers.length; i++) {
                $scope.pics.markers[i].setMap(null);
              }
              $scope.pics.markers = [];
            }).
            error(function(data, status, headers, config) {
              $scope.pics.messagetitle = "Error";
              $scope.pics.messagebody = "Picture can't be deleted"
              if(data != "") {
                $scope.pics.messagebody += "<br /><br /><pre class='prettyprint'><code class='language-xml'>" + data.encodeHTML() + "</pre></code>";
              }
              $('#message').modal({show: true});
            });
            thumbnailHeaders = {};
            for (var header in $scope.pics.pictures[index]["delete_request_thumbnail_headers"]) {
              thumbnailHeaders[header] = $scope.pics.pictures[index]["delete_request_thumbnail_headers"][header][0];
            }
            $http({
              url: $scope.pics.pictures[index]["delete_request_thumbnail_url"],
              method: "DELETE",
              headers: thumbnailHeaders
            }).
              success(function(data, status, headers, config) {
              }).
              error(function(data, status, headers, config) {
                $scope.pics.messagetitle = "Error";
                $scope.pics.messagebody = "Thumbnail can't be deleted"
                if(data != "") {
                  $scope.pics.messagebody += "<br /><br /><pre class='prettyprint'><code class='language-xml'>" + data.encodeHTML() + "</pre></code>";
                }
                $('#message').modal({show: true});
              });
        };
      }],
      controllerAs: "showCtrl"
    };
  });

  // Display the map
  app.directive("picsBigmap", function() {
    return {
      restrict: 'E',
      templateUrl: "app/html/pics-bigmap.html",
      controller: ['$http', '$scope', 'picsService', function($http, $scope, picsService) {
        $scope.pics = picsService;
        this.displayBigMap = function() {
          if($(window).height() > $(window).width()) {
            $("#bigmap").height($(window).width());
          } else {
            $("#bigmap").height($(window).height() * 0.8);
          }
          var bigmapOptions = {
            zoom: 2,
            center: new google.maps.LatLng(35, -100),
            mapTypeId: google.maps.MapTypeId.ROADMAP
          };
          $scope.pics.bigmap = new google.maps.Map(document.getElementById('bigmap'),
          bigmapOptions);
          var bounds = {
            north: 50,
            south: 20,
            east: -70,
            west: -130
          };
          $scope.pics.rectangle = new google.maps.Rectangle({
            bounds: bounds,
            editable: true,
            draggable: true
          });
          $scope.pics.rectangle.setMap($scope.pics.bigmap);
          $scope.pics.rectangle.addListener('bounds_changed', this.updateRectanglePosition);
          google.maps.event.trigger($scope.pics.bigmap, 'resize');
        };
        this.updateRectanglePosition = function(event) {
          var ne = $scope.pics.rectangle.getBounds().getNorthEast();
          var sw = $scope.pics.rectangle.getBounds().getSouthWest();

          var contentString = '<b>Rectangle moved.</b><br>' +
              'New north-east corner: ' + ne.lat() + ', ' + ne.lng() + '<br>' +
              'New south-west corner: ' + sw.lat() + ', ' + sw.lng();

          console.log(contentString);
        }
      }],
      link: function(scope, elem, attrs) {
        scope.bigmapCtrl.displayBigMap();
      },
      controllerAs: "bigmapCtrl"
    };
  });

  app.directive("picsMessage", function() {
    return {
      restrict: 'E',
      templateUrl: "app/html/pics-message.html"
    };
  });

  app.directive("picsInfos", function() {
    return {
      restrict: 'E',
      templateUrl: "app/html/pics-infos.html"
    };
  });
})();
