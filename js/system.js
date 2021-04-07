const time = new EventSource('/time');
time.addEventListener('time', (e) => {
    document.getElementById("time").innerHTML = e.data;

}, false);
const overview = new EventSource('/overview');
overview.addEventListener('overview', (e) => {
    const overviewData = e.data.split(";");
    document.getElementById("production").innerHTML = overviewData[0];
    document.getElementById("downtime").innerHTML = overviewData[1];
    document.getElementById("offline").innerHTML = overviewData[2];

}, false);
const workplaces = new EventSource('/workplaces');
workplaces.addEventListener('workplaces', (e) => {
    const overviewData = e.data.split(";");
    document.getElementById(overviewData[0]).innerHTML = overviewData[1];
    document.getElementById(overviewData[0]).style.background = overviewData[2]
}, false);