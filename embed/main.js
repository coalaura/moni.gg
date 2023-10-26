(function ($) {
	const widths = [
		{ min: 1024, days: 90 },
		{ min: 600, days: 60 },
		{ min: 0, days: 30 }
	];

	const colors = [
		{min: 60, color: "#ff2600"},
		{min: 30, color: "#eba019"},
		{min: 10, color: "#e2c719"},
		{min: 1, color: "#83da56"},
		{min: 0, color: "#05f4a7"}
	];

	const rectWidth = 3,
		rectPadding = 2;

	function getDays() {
		const width = window.innerWidth;

		return widths.find(w => width >= w.min).days;
	}

	function getColor(value) {
		return colors.find(c => value >= c.min).color;
	}

	function getViewBox() {
		const days = getDays();

		const viewBox = [];

		if (days === 90) {
			viewBox.push(0);
		} else {
			const offset = 90 - days;

			viewBox.push((offset * rectWidth) + (rectPadding * (offset))); // x origin
		}

		viewBox.push(0); // y origin

		viewBox.push((rectWidth * days) + (rectPadding * (days - 1))); // svg width
		viewBox.push(34); // svg height

		return viewBox.join(' ');
	}

	function getTitle(date, downtime) {
		const hours = Math.floor(downtime / 60);
		downtime = downtime % 60;

		const time = downtime > 0 ? `${hours.toString().padStart(2, "0")}:${downtime.toString().padStart(2, "0")}` : "No downtime";

		return date.format("ddd, MMM Do") + ": " + time;
	}

	function renderSVG(history = {}) {
		const rects = [];

		for (let day = 0; day < 90; day++) {
			const date = moment().utc().subtract(day, 'days'),
				index = date.format("YYYY-MM-DD"),
				downtime = history.downtimes[index] || 0,
				x = (90 - day - 1) * (rectWidth + rectPadding);

			rects.push(`<rect height="34" width="${rectWidth}" x="${x}" y="0" fill="${getColor(downtime)}" class="slice" title="${getTitle(date, downtime)}"></rect>`);
		}

		const days = getDays(),
			total = Object.values(history.downtimes).reduce((a, b) => a + b, 0),
			uptime = ((1 - (total / (90 * 24 * 60))) * 100).toFixed(2).replace(/\.0+$|(?<=\.\d)0+$/gm, "");

		const legend = [
			`<div class="legend">`,
			`<div class="item">${days} days ago</div>`,
			`<div class="spacer"></div>`,
			`<div class="item uptime">${uptime} % <span class="no-mobile">uptime</span></div>`,
			`<div class="spacer"></div>`,
			`<div class="item">Today</div>`
		];

		return {
			svg: `<svg preserveAspectRatio="none" height="34" viewBox="${getViewBox()}">${rects.join('')}</svg>`,
			legend: legend.join('')
		};
	}

	function getHeader(data) {
		const all = Object.values(data.data).length;

		let text = "All Systems Operational",
			image = "available.png",
			color = "#05f4a7";

		if (data.down === all) {
			text = "Full Outage";
			image = "full.png";
			color = "#ff2600";
		} else if (data.down > 1) {
			text = "Major Outage";
			image = "major.png";
			color = "#eba019";
		} else if (data.down === 1) {
			text = "Partial Outage";
			image = "partial.png";
			color = "#e2c719";
		}

		return {
			text: text,
			image: image,
			color: color
		};
	}

	function update() {
		$.get("status.json?_=" + Date.now(), function (data) {
			const header = getHeader(data);

			$("#status span").text(header.text);
			$("#status img").attr("src", header.image);
			$("#status").css("background-color", header.color);

			$("#services").empty();

			$.each(data.data, function (name, status) {
				const svg = renderSVG(status.history);

				const html = [
					`<div class="service ${status.operational ? "up" : "down"}">`,
					`<div class="header">`,
					`<span class="name">${name} <sup>${status.type}</sup></span>`,
					`<span class="status">${status.operational ? "Operational" : "Outage"}</span>`,
					`</div>`,
					`<div class="body">`,
					svg.svg,
					svg.legend,
					`</div>`,
					`</div>`
				];

				$("#services").append(html.join(''));
			});

			const date = moment(data.time * 1000);

			$("#time").attr("title", "Status was last updated " + date.from());
			$("#time").text(date.format('dddd, MMMM Do YYYY, h:mm:ss a'));
		});
	}

	update();

	setInterval(update, 15000);

	let timeout;

	$(window).on("resize", function () {
		clearTimeout(timeout);

		timeout = setTimeout(function () {
			update();
		}, 250);
	});
})($);