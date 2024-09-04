# Nick Slam

### Build:
```
make
```

### Example Config:
```json
{
  "services": [
    {
      "name": "nickslam1",
      "namespace": "rdk",
      "type": "slam",
      "model": "ncs:slam:nickslam",
      "attributes": {
        "big": false
      }
    }
  ],
  "modules": [
    {
      "type": "local",
      "name": "ncs-nickslam",
      "executable_path": "/home/user/nickslam"
    }
  ]
}
```
