# Simple Docker volume backup tool

## Usage

```bash
> volumesaver -c config.json
```

or assuming there is a config.json file

```bash
> volumesaver
```

## Config.json

```json
{
    "container_name": "<container name>",
    "backup_dir": "<destination folder>",
    "temp_dir": "<temp folder>",
    "volumes": [
      {
        "container_path": "<volume path #1 within the container>"
      },
        ...
	  {
        "container_path": "<volume path #n within the container>"
      }
    ]
  }
```