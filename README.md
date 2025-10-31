# Aurapack

I started automating my home. Lights, sensors, thermostats, locks, a media server. At first it worked well enough. I wasn’t chasing comfort. Most days it is faster to flip a switch than to speak to a device.

Still, some things are not about comfort. Turning a light on is easy, but changing its color or syncing it with movement is not. Some routines are simple by hand, others make more sense when they happen on their own.

It didn’t take long for the cracks to show. A few times, late at night, I asked to dim a lamp and got a Korean playlist at full volume instead. Then came the shouting match with the assistant, trying to stop the music before waking the whole house. Sometimes the second command would turn on another light, or all of them, or switch something off that shouldn’t be touched.

After wiring most of the house the feeling turned mixed. Part of me wanted strong analog switches that never fail. The rest of me liked the orchestration but kept hitting glitches. Voice recognition feels worse today than a few years ago in both the Google and AWS ecosystems.

Privacy adds weight. Cheap devices are easy to buy and hard to trust. I do not know where the data goes. I do not know who can see it.

The original goal was simple. Schedule lights. Trigger lamps on motion. Bump the heating a few degrees if I get home earlier. Small decisions that should run close to where they matter.

This project is a reply to that experience. A local first assistant that can think nearby and speak without a round trip to someone else’s server. Speech on a Raspberry Pi. Reasoning on a home server. Sensors and actuators over MQTT. Remote help only when it makes sense.