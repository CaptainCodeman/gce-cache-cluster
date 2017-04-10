# Example

Example project that will serve files from a cloud storage bucket
but caching them in groupcache. Imagine this being part of an image
handling system that also did image processing - you could have a
cache layer for source images to avoid hitting GCS each time and
another cache layer for resized images to avoid re-processing them
on each request, and each source-load and image-resize would only
happen on a single node, with other nodes waiting for completion.

## Installing

This example is based on deployment to Docker Optimized GCE. You
should change refereces to project, zones and buckets to match your
project.

Install and setup:

* Docker
* gcloud SDK
* govendor

Run `govendor install` to install the vendored dependencies required
to build the docker container.

## Running

These are examples only to fire up some instances:

Set gcloud project:

    gcloud config set project my-project

Build docker image and push to gcloud registry:

    docker build -t my-project-test .
    docker tag my-project-test gcr.io/my-project/test
    gcloud docker -- push gcr.io/my-projecte/test

Create a GCE instance using this image:

    gcloud compute instances create my-project-test-1 \
        --image-family cos-stable \
        --image-project cos-cloud \
        --zone us-central1-b \
        --machine-type f1-micro \
        --tags http-server \
        --metadata service=test \
        --metadata-from-file user-data=cloud-init.txt

See cloud-init.txt for service and networking settings.

