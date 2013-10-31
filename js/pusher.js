function Pusher(options) {
  var that = this;
	// socket url
	that.url = options.url;
	// heartbeat interval
	that.heartbeat = null;
	// session id
	that.id = null;
	// min backoff
	that.minBackoff = 500;
	// max backoff
	that.maxBackoff = 30000;
	// exponential backoff
	that.backoff = that.minBackoff;
	// interval that sends heartbeats
	that.heartbeater = null;
	// socket
	that.socket = null;
	// reconnect timeout
	that.reconnector = null;
	// buffer while closed
	that.buffer = [];
	// callbacks
	that.callbacks = {};
	/*
	 * set up the socket
	 */
	that.connect = function() {
	  var url = that.url;
		if (that.id != null) {
		  url += '?session_id=' + that.id;
		}
		that.socket = new WebSocket(url);
		that.socket.onopen = function() {
			that.lastHeartbeatReceived = new Date();
			that.backoff = that.minBackoff;
			while (that.buffer.length > 0) {
			  that.send(that.buffer.shift());
			}
		};
		that.socket.onerror = function() {
			that.close();
		};
		that.socket.onmessage = function(message) {
			var msg = JSON.parse(message.data);
			that.handleMessage(msg);
		};
		that.socket.onclose = function() {
			that.close();
		};
	}
	/*
	 * handle incoming messages
	 */
	that.handleMessage = function(msg) {
		if (msg.Type == "Welcome") {
			that.heartbeat = msg.Welcome.Heartbeat;
			that.id = msg.Welcome.Id;
			if (that.heartbeater != null) {
				clearInterval(that.heartbeater);
			}
			that.heartbeater = setInterval(function() {
				if (new Date().getTime() - that.lastHeartbeatReceived.getTime() > that.heartbeat) {
					that.close()
				} else {
					that.send({
						Type: "Heartbeat",
					});
				}
			}, that.heartbeat / 2);
		} else if (msg.Type == "Heartbeat") {
			that.lastHeartbeatReceived = new Date();
		} else if (msg.Type == "Message") {
			var callbacks = that.callbacks[msg.URI];
			if (callbacks != null) {
			  for (var callback in callbacks) {
				  callbacks[callback](msg.Data);
				}
			}
		}
	};
	that.emit = function(uri, data) {
	  that.send({
		  Type: 'Message',
			URI: uri,
			Data: data,
		});
	};
	that.on = function(uri, callback) {
	  if (that.callbacks[uri] == null) {
		  that.callbacks[uri] = {};
			that.send({
				Type: 'Subscribe',
				URI: uri,
			});
		}
		that.callbacks[uri][callback] = callback;
	};
	that.off = function(uri, callback) {
	  delete(that.callbacks[uri][callback]);
		var left = 0;
	  for (var callback in that.callbacks[uri]) {
		  left++;
		}
		if (left == 0) {
		  that.send({
			  Type: 'Unsubscribe',
				URI: uri,
			});
			delete(that.callbacks[uri]);
		}
	};
	/*
	 * close
	 */
	that.close = function() {
	  clearInterval(that.heartbeater);
		that.heartbeater = null;
		that.socket.close();
		if (that.backoff < that.maxBackoff) {
		  that.backoff *= 2;
		}
		if (that.reconnector != null) {
		  clearTimeout(that.reconnector);
		}
		that.reconnector = setTimeout(that.connect, that.backoff);
	};
	/*
	 * send a JSON encoded obj
	 */
	that.send = function(obj) {
		if (that.socket.readyState == 1) {
			that.socket.send(JSON.stringify(obj));
		} else {
		  that.buffer.push(obj);
		}
	};
	that.connect();
  return that;
}
