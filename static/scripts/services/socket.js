'use strict';

angular.module('chatWebApp')
    .factory('socket', function (socketFactory) {
        var socket = socketFactory();
        socket.forward('error');
        return socket;
    });