# Local TLE Database and Server-Client Process

**Basic file only - a work in progress**

This repo provides the code to operate a local TLE mirror from the SpaceTrack service.
It will grab the current TLEs at the specified cadence, and service requests for specific TLEs in a similar way to the existing SpaceTrack API.

## Initial setup
Clone the repo to the desired location
Edit the config options as desired - see below for details about what can be done
Set up your SpaceTrack credentials
Set up your database
### Install go (server) module
- Navigate to local_tles/server directory
```bash
go build # creates an executable for the server
go install

# add the executable directory to $PATH if necessary

echo -e "\n# local_tles setup and config" >> ~/.bashrc

executable_file=`go list -f '{{.Target}}'`
executable_dir=$(dirname "$executable_file")
if [[ ":$PATH:" == *":$executable_dir:"* ]]; then
   echo "$executable_dir already in \$PATH" 
else
    echo "export PATH="'$PATH'":$executable_file" >> ~/.bashrc
fi

# set up environment variables
echo "LTLE_CONFIG=`pwd`/config/; export LTLE_CONFIG" >> ~/.bashrc

source ~/.bashrc # to start using now in current bash session
```

### Install python (client) package
- Navigate to local_tles/client directory
- pip install -e .
- Module can be used in python file or interpreter with `import ltle`

### Set up systemd service to run server 
TODO - install the service to make it run at startup

## Dependencies


## Config Options
An example config file as follows:
```
READ_RATE 1
FIRST_DATE 2024-01-01
DATASTORE /path_to_repo/local_tles/data/st_files/
DATABASE /path_to_repo/local_tles/data/db/tle_info.db
PORT 3333
ACCESS_ST 1
```

`READ_RATE` is the cadence (in hours) that space-track.com will be queried to get data for required days.
`FIRST_DATE` is the earliest date the program will query from space-track.com.
`DATASTORE` and `DATABASE` specify the file locations to store the queried data. These locations do not need to be in subdirectories of your local copy of this repository.
`PORT` determines the port to run the local server on.
`ACCESS_ST` is a boolean flag to allow running of the server *without* querying space-track.com. Intended for testing, to avoid being banned from too many queries.

Note that currently, the state of this program only allows for manual triggering of queries, and the DB is not operational. space-track.com results are stored in `DATASTORE` and can still be accessed and read by manually running python commands as per the client package


## Credentials
The credential file consists of two items, as per your personal space-track.com credentials:
```
username user.name@email.com
password p4SSw0rd
```

Ensure you do not commit these details to any repositories.

## TODO List
Currently not much is operational.
Things that work:
- Querying SpaceTrack for the current day's TLEs in JSON format
- Or for the date string in argv[1]
- Querying through the python client, which starts and runs the program as needed

Things that don't / aren't implemented:
- More complicated SpaceTrack queries
- Scheduled queries at the specified cadence
- Adding the queried data to the database
- Exposing an API for clients to query
- This includes ip/port for local network queries to different machines
- This should be a mostly-drop-in replacement for the python spacetrack package
- Maybe a more disk efficient CSV format instead (of JSON)?
- Immediate querying and response if able via client requests while the server is background running
- Queue to add requests if not able, with duration until complete
- Figure out what data to keep in the DB
