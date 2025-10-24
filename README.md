# Scan to Nextcloud

A simple Go app that uploads any files in a given set of folders to Nextcloud and deletes the source files.

The main goal of this app is to let a scanner upload files to Nextcloud, essentially the scanner drops the file
into a network share, and this app will upload it to Nextcloud and delete it from the network share.

## Quickstart

I run this as a Kubernetes CronJob, so the Docker image doesn't have any scheduling features, it just runs once
and then exits. You could just grab the `scan-to-nextcloud` executable file from a release and set it up with
a Linux cron job if you aren't running a Kubernetes cluster.

**Note: The configuration file contains sensitive information, make sure you store it in a Kubernetes secret or adjust the file permissions wherever you are storing it.**

### Environment Variables

- `STNC_CONFIG_PATH`: The absolute path to the configuration file (defaults to `./config.yaml`, which is `/app/config.yaml` in the Docker image)

### Configuration File

```yaml
base_path: /mnt/scans # The absolute path to the input root directory
nextcloud_url: https://cloud.my.domain # The base URL for the Nextcloud instance, no trailing slash
users:
  - username: Bob # WebDAV compatible Nextcloud username
    api_key: xxxxx # Create an API key for the user
    input_folder: Bobs_Folder # Relative path from base bath, resolves to `/mnt/scans/Bobs_Folder` in this example
    output_folder: Scans # Put scanned files into a folder called "Scans" in the users Nextcloud account
```
