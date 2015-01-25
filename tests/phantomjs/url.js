console.log('Trying to upload a cute corgi...');

var page = require('webpage').create();
var corgi = 'http://i.imgur.com/zJqtNL0.jpg';
var url = 'http://127.0.0.1:8080/url';

page.open(url, 'POST', 'image=' + encodeURIComponent(corgi), function(status) {
    console.log('Status: ' + status);

    if(status === 200) {
        phantom.exit(0);
    } else {
        phantom.exit(1);
    }
});
