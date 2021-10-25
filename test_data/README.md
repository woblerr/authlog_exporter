# Custom Maxmind database for tests

Create a custom Maxmind database for tests.

## Using docker

Simple
```bash
make docker-build-test-db
```
or
```bash
cd ./test_data
docker build -f Dockerfile.testdb -t auth_exporter_build_test_db .
docker run -d --name=auth_exporter_build_test_db auth_exporter_build_test_db
docker cp auth_exporter_build_test_db:/db/geolite2_test.mmdb ./
docker rm -f auth_exporter_build_test_db
```
