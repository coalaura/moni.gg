(function ($) {
    const page = $("#page");

    function renderHistoric(historic) {
        if (!historic) {
            historic = {};
        }

        const max = 60;

        const now = Math.floor((Date.now() / 1000) / 600);

        let hours = {};

        for (let min = now - (144 * 5); min < now; min++) {
            const hour = Math.floor(min / 6);

            if (!(hour in hours)) {
                hours[hour] = 0;
            }

            const value = min in historic ? historic[min] : 0;

            hours[hour] += value;
        }

        let html = [];

        let total = 0;

        $.each(hours, (hour, amount) => {
            const date = moment(hour * 60 * 60 * 1000);

            total += Math.min(amount, max);

            const hsl = amount > 0 ? Math.ceil((1 - (Math.min(amount * 4, max) / max)) * 120) : 120;

            html.push(`<div class="slice" style="background: hsl(${hsl}, 100%, 80%)" title="${date.format('MMMM Do YYYY, ha')}: ${amount > 0 ? amount + " minute(s)" : "No"} downtime."></div>`);
        });

        return {
            html: html.join(""),
            uptime: Math.floor((1 - (total / (120 * max))) * 100 * 100) / 100
        };
    }

    function getStatusData(data) {
        const all = Object.values(data.data).length;

        if (data.down === all) {
            return {
                title: "Everything is broken",
                image: "full.png"
            };
        } else if (data.down > 1) {
            return {
                title: "A bunch of things are not working",
                image: "major.png"
            };
        }

        return {
            title: "Something seems wrong",
            image: "partial.png"
        };
    }

    function update() {
        $.get('status.json?_=' + Date.now(), function (data) {
            page.html('');

            if (data.down === 0) {
                page.append('<div id="status" class="up">Yep, everything is fine <img class="main-status" src="available.png" /></div>');
            } else {
                const status = getStatusData(data);

                page.append('<div id="status" class="down">' + status.title + ' <img class="main-status" src="' + status.image + '" /></div>');
            }

            page.append('<div id="services"></div>');

            $.each(data.data, function (name, status) {
                const since = status.status ? moment(status.status * 1000) : false,
                    historic = renderHistoric(status.historic);

                $('#services').append('<div class="service ' + (status.status ? 'down' : 'up') + '" title="' + (status.status ? 'Service has first been unavailable ' + since.format('dddd, MMMM Do YYYY, h:mm:ss a') + '.' : 'Service is available.') + '"><span class="name">' + name + ' <sup>' + status.type + '</sup></span><span class="status-msg">' + (status.status ? 'Unavailable since ' + since.fromNow(true) : 'Available') + ' (' + historic.uptime + '% uptime)</span><div class="historic">' + historic.html + '</div></div>');
            });

            const date = moment(data.time * 1000);

            page.append('<div id="time" title="Status was last updated ' + date.from() + '.">' + date.format('dddd, MMMM Do YYYY, h:mm:ss a') + '</div>');
        });
    }

    update();

    //setInterval(update, 5000);
})($);