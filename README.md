## Backend for personal plant monitoring project
This is a backend service for handling MQTT data coming from sensors attached to plants. 

It can store the incoming readings in a Timescale database, as well as serve said readings either from a select timeframe with a REST API, or in realtime using websockets.

Since this is a personal project, the code is provided as is, mainly for portofolio purposes.


### Considerations
A quick list of considerations made during the design of the project

- The device configurations are stored in a json file. Because this is a small-scale project, designed to be run without any support, etc... I believed it would be good to allow easy editing of configuration using a text editor, rather than needing a database query. JSON was chosen because it can be parsed easily by Golangs built-in utilities.

- Initially I wanted the database to be a separate goroutine, like the mqtt and http modules. However it proved more convenient to simply pass a reference to the db manager around, since it is designed to be thread safe anyway. Since data needs to flow both in and out of the db manager at a number of locations in the code, a non-blocking way of bidirectional communication with both mqtt and http would add too much code complexity.