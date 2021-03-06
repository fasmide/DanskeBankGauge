package danskebank

// JSTemplate is used to evaluate the javascript sealer from danskebank
const JSTemplate = `
let handler = {
    get: function(target, name) {
        // this is to allow for the random number value they provide 
        // i beleive this is because they are trying not to clash with other variables
        if (name.length == 33 && target[name] == undefined) {
            console.error("created", name);
            target[name] = {};
        }

        // lets see if there should appear any thing we forgot to implement in our fake window object
        if (target[name] == undefined && name != "undefined") {
            console.error(name, "was not set")
        }
        return target[name];
    }
}
window = {
    atob: require('atob'),
    btoa: require('btoa'),
    Infinity: Infinity,
    ActiveXObject: {},
    Array: Array,
    Date: Date,
    Int16Array: Int16Array,
    JSON: JSON,
    Math: Math,
    RegExp: RegExp,
    String: String,
    Uint8Array: Uint8Array,
    addEventListener: function() {},
    document: {
        createElement: function() {
            console.error("createElement", arguments)
            return {
                getElementsByTagName: function() {
                    console.error("getElementsByTagName", arguments)
                    return new Proxy({}, handler)
                }
            }
        },
    addEventListener: function() {},
    },
    location: {},
    navigator: {platform: "linux" },
    sessionStorage: {setItem: function() {console.error("setItem", arguments)}},
    undefined: undefined,
    setTimeout: setTimeout,
    parseInt: parseInt,
    clearTimeout: function() {console.error("gonna clear...", arguments); }
};

window.top = window;

window = new Proxy(window, handler)
window.getDeviceInformationString = function(cb) {cb("https://github.com/fasmide/DanskeBankGauge");}


eval({{ .Signer }});

performLogonServiceCode_v2("{{ .SSN }}", "{{ .SC }}", function(package) {
    console.log(JSON.stringify(package));
    process.exit(0)
}, function() {
    console.log("failed")
    process.exit(1)
});
`
