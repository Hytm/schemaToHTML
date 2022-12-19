# Cockroach DB schema in HTML
Small tool to generate movable div to visualize a schema from a database in Cockroach.

The tool collects tables definitions, Primary Keys, and Foreign Keys.
It is based on the work from @Nelms on Code Pen.

![alt text](https://ibb.co/CVQF3jk "Example file")

# how to use the app
Only use 2 parameters

| flag          | usage                                                    | Optional      |
| ------------- |:--------------------------------------------------------:|:-------------:|
| -c            | Connection string with username, password, and database  | No            |
| -f            | Shows information for type, nullable, and default value  | Yes           |
