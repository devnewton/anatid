var http = require('http');
var express = require("express");
var RED = require("node-red");

var app = express();
var server = http.createServer(app);

var settings = {
    httpAdminRoot: "/",
    httpNodeRoot: "/",
    adminAuth: {
        type: "credentials",
        users: [],
        default: {
            permissions: "read"
        }
    },
    userDir: "$HOME/.config/anatid",
    flowFile: "flows_anatid.json",
    functionGlobalContext: {}
};

RED.init(server, settings);
app.use(settings.httpAdminRoot, RED.httpAdmin);
app.use(settings.httpNodeRoot, RED.httpNode);
server.listen(8000);
RED.start();
