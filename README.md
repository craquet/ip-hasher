# ip-hasher

This service tails the `access.log` of an Apache2 server and outputs an `access_hashed.log` containing the same log lines, with the IPs hashed using sha256.

Can be deployed using Docker with the provided `Dockerfile` and `docker-compose.yml`

### Environment

| Variable     | Description                       |
|--------------|-----------------------------------|
| FILENAME_IN  | Filename of the input access log  |
| FILENAME_OUT | Filename of the output access log |

### Docker Notes

Input files must be mounted at `/app/logs`. Output files must be mounted at `/app/out`.

### Technical Details

This service opens and then continuously tails the provided access.log (polling). All new lines will have their IP hashed and will then be appended to the output file. It also stats the file every second, to catch inode changes that signal a log rotation. In case of a log rotation, the input file is closed, the output file is truncated and the program will restart.

A more ressource-friendly implementation could be using filesystem signals (fsnotify), but this might be problematic with Docker.
