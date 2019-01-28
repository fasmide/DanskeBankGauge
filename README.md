# Readme

This project is about making a gauge sitting on my fridge displaying the amount of money is available on our households food bank account
We usually make the bank auto transfer a fixed amount to this account and then use it thoughout the month to buy everyday stuff - but somehow we forget to check if there is any funds available on it - this should fix that by bringing the data directly to the kitchen

## Whats in here

* Some arduino firmware for driving a led strip and the gauge it self
* A golang lib for talking to the mobile api
* A golang daemon that provides a much simpler api to the arduino firmware

## Firmware

The gauge is just PWM driven with a resistor in series and the led strip is some ws2812 strip i had laying around.
It pulls a http endpoint providing some simple json to figure out how much money is on the account, its not talking directly to danske bank's api as it proved somewhat challanging and a lot easier to manage on a pc. 

## Daemon

This is just a real simple golang http server which provides a `/balance` endpoint and either asks the bank for data or returns a cached version

## Danske Bank's Mobileapi

We are using Danske bank's mobileapi to talk to, they do not provide any documentation on this and i had to reverse engineer the whole thing by taking a good hard look at their mobile app and how it communicates with it.

Danske bank currently have 2 mobile applications available - here we are working with the new app's api - the other app (with a different api) have been reversed by others but it seems like danske bank maybe discontinueing their old app when their new one is ready.

Its pretty complicated to actually use, the mobile app adds some `x-ibm-client-id` and `x-ibm-client-secret` headers to every request, these are just hardcoded values which can be found in the mobile app, but if you dont set them, their service just tell you "Unauthorized client".

The logon method requires users to post a `LogonPackage` which i have no idea how to create so im just doing it the way their mobile app does it. Every time a user wants to logon the app fetches some obfuscated javascript (which btw seems to change every time you fetch it), it evaluates the javascript and then calls a global function presented by the fetched javascript called `performLogonServiceCode_v2(cpr, sc, success, error)`. This function takes the CPR and SC pinnumber and if all goes well call's the success callback with a `LogonPackage` - once this payload is posted to their logon endpoint its just a matter of passing around an auth token with the `Authorization` http header.

Once you get there, the api is some simple http endpoints which you post some json payloads to and receive various json responses from - i'm only interrested in their `account/list` endpoint as this returns the balance data i need for this project. But i would think it does provide everything thats visible in the mobile app.

On second thought it would have been a lot easier just waiting on their open banking initiative - or maybe even using a third party provider like NordicAPIGateway - these guys provides a single api that covers nearly all the nordic banks - This was just such a good challange to figure out how they do it in the mobile app.

