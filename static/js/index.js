var time = 0;
var timer = 0;



function getDataAndShowGraph(url, elementId, label) {
  $.get(url, {}, function(result) {
    var data = JSON.parse(result);
    var max_val = 0.0;
    data.data.forEach(function(val) {
      if (parseFloat(val) > max_val) {
        max_val = parseFloat(val);
      }
    });

    var ctx = document.getElementById(elementId);
    var myLineChart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: data.time,
        datasets: [{
          label: label,
          lineTension: 0.3,
          backgroundColor: "rgba(2,117,216,0.2)",
          borderColor: "rgba(2,117,216,1)",
          pointRadius: 4,
          pointBackgroundColor: "rgba(2,117,216,1)",
          pointBorderColor: "rgba(255,255,255,0.8)",
          pointHoverRadius: 4,
          pointHoverBackgroundColor: "rgba(2,117,216,1)",
          pointHitRadius: 3,
          pointBorderWidth: 2,
          data: data.data
        }],
      },
      options: {
        scales: {
          xAxes: [{
            time: {
              unit: 'time'
            },
            gridLines: {
              display: false
            },
            ticks: {
              maxTicksLimit: 7
            }
          }],
          yAxes: [{
            ticks: {
              min: 0,
              max: max_val,
              maxTicksLimit: 5
            },
            gridLines: {
              color: "rgba(0, 0, 0, .125)",
            }
          }],
        },
        legend: {
          display: false
        }
      }
    });


  });

}

function updateGraphs() {
  getDataAndShowGraph("/mainData/getGraphHumData", "mainAreaHumidityChart", "влажность");
  getDataAndShowGraph("/mainData/getGraphTempData", "mainAreaTemperatureChart", "температура");
}




function DecTime() { 
    if (time > 0) {
        time--;
        document.getElementById("clockdiv").innerHTML = 'до возвращения в автоматический режим: ' + time + ' сек';
    }
    else {
        document.getElementById("clockdiv").innerHTML = '';
        clearInterval(timer);
        timer = 0;
    }
}

function restartCountdown(seconds) { 
  if (timer != 0) { 
      clearInterval(timer); 
  } 
  time = seconds;  
  document.getElementById("clockdiv").innerHTML = 'до возвращения в автоматический режим: ' + time + ' сек';
  timer = setInterval(function() { 
      DecTime(); 
  }, 1000); 
} 

function sendRequestToMain(request) {
  $.get("/mainData/"+request, {}, function(result) {
    if (result != "") { 
      restartCountdown(result);
    }
  });
}

function initApp() {
  loadMainStatus();
  document.getElementById("camera_photo").src = "img.jpg?"+Math.random();
  setInterval(function() {
    loadMainStatus();
  }, 1000);
  updateGraphs();
  setInterval(function() {
    updateGraphs();
  }, 60000);
}

function hideAllPages() {
  document.getElementById("pageAreaMain").style.display = "none";
  document.getElementById("pageAreaGrowing").style.display = "none";
  document.getElementById("pageAreaSystem").style.display = "none";
  document.getElementById("pageAreaInfo").style.display = "none";
}

function showPage(pagename) {
  document.getElementById(pagename).style.display = "block";
}

function showMainPage() { 
  hideAllPages();
  showPage("pageAreaMain");
} 
function showGrowingSettingsPage() { 
  hideAllPages();
  showPage("pageAreaGrowing");
} 
function showSystemSettingsPage(){ 
  hideAllPages();
  showPage("pageAreaSystem");  
} 
function shutdownComputer() { sendRequestToMain("shutdown"); } 
function toggleLight() { sendRequestToMain("toggleLight"); } 
function togglePump() { sendRequestToMain("togglePump"); } 
function toggleFan() { sendRequestToMain("toggleFan"); } 
function toggleChiller() { sendRequestToMain("toggleChiller"); } 
function updatePhoto() { 
  $.get("/mainData/updatePhoto", {}, function(result) {
    document.getElementById("camera_photo").src = "";
    document.getElementById("camera_photo").src = "img.jpg?"+Math.random();
  });
} 

function saveGrowingSettings() {
  var gSet = {
    valLightOnTime: document.getElementById("valLightOnTime").value,
    valLightOffTime: document.getElementById("valLightOffTime").value,
    valPumpOnTime: document.getElementById("valPumpOnTime").value,
    valPumpPauseTime: document.getElementById("valPumpPauseTime").value,
    valFanOnTime: document.getElementById("valFanOnTime").value,
    valFanPauseTime: document.getElementById("valFanPauseTime").value,
    valChillerOnThreshold: document.getElementById("valChillerOnThreshold").value
  };
  if (gSet.valLightOnTime == "") {
    gSet.valLightOnTime = document.getElementById("valLightOnTime").placeholder;
  }
  if (gSet.valLightOffTime == "") {
    gSet.valLightOffTime = document.getElementById("valLightOffTime").placeholder;
  }
  if (gSet.valPumpOnTime == "") {
    gSet.valPumpOnTime = document.getElementById("valPumpOnTime").placeholder;
  }
  if (gSet.valPumpPauseTime == "") {
    gSet.valPumpPauseTime = document.getElementById("valPumpPauseTime").placeholder;
  }
  if (gSet.valFanOnTime == "") {
    gSet.valFanOnTime = document.getElementById("valFanOnTime").placeholder;
  }
  if (gSet.valFanPauseTime == "") {
    gSet.valFanPauseTime = document.getElementById("valFanPauseTime").placeholder;
  }
  if (gSet.valChillerOnThreshold == "") {
    gSet.valChillerOnThreshold = document.getElementById("valChillerOnThreshold").placeholder;
  }
  //alert(JSON.stringify(gSet));
  $.ajax({
    type : 'POST',
    url: 'growingSettings',
    data: JSON.stringify(gSet),
    success: function () {
      hideAllPages();
      document.getElementById("info_text").innerHTML = "<br><div class='row'><div class='col'><p><label>Настройки успешно сохранены</label></p></div></div>";
      showPage("pageAreaInfo"); 
    }
  });
  /*$.post('/growingSettings', {data : JSON.stringify(gSet)}, function (data) {
      alert(data); 
  });*/
}

function loadGrowingSettings() {
  $.get("growingSettings", {}, function(result) {
    var data = JSON.parse(result);
    document.getElementById("valLightOnTime").value = data.valLightOnTime;         
    document.getElementById("valLightOffTime").value = data.valLightOffTime; 
    document.getElementById("valPumpOnTime").value = data.valPumpOnTime;
    document.getElementById("valChillerOnThreshold").value = data.valChillerOnThreshold;
    document.getElementById("valPumpPauseTime").value = data.valPumpPauseTime;
    document.getElementById("valFanOnTime").value = data.valFanOnTime;
    document.getElementById("valFanPauseTime").value = data.valFanPauseTime;
    showGrowingSettingsPage();
  });
  
}

function IsContainChars(str, chars) {
  for (var i=0; i < str.length; i++) {
    for (var j=0; j < chars.length; j++) {
      if (str[i] == chars[j]) {
        return true;
      }
    }
  }
  return false;
}

function isValidSSID(str) { 
  if (str.length < 8 || str.length > 32) {
    return false
  }
  if (IsContainChars(str, "~`@#$%^&*()=+/\\|[]{}:;\"<>_,?йцукенгшщзхъфывапролджэячсмитьбюЙЦУКЕНГШЩЗФЫВАПРОЛДЖЭЯЧСМИТЬБЮ")) {
    return false;
  }
  return true;

}
function isValidWPA(str) { return /^[\u0020-\u007e\u00a0-\u00ff]*$/.test(str); }


function saveSystemSettings() {
  var gSet = {
    valSsid: document.getElementById("valSsid").value,
    valPass: document.getElementById("valPass").value,
    valAPName: document.getElementById("valAPName").value
  };

  if (!isValidSSID(gSet.valAPName)) {
    alert("Ошибка в имени PFC (точки доступа), длина 8..32 символов, может содержать a..z, A..Z, 0..9, символы . -");
    return 0;
  }
  
  if (gSet.valSsid == "") {
    alert("Не задано имя сети!");
    return 0;
    gSet.valSsid = document.getElementById("valSsid").placeholder;
  }

  if (gSet.valAPName == "") {
    gSet.valAPName = document.getElementById("valAPName").placeholder;
  }
  //alert(JSON.stringify(gSet));
  $.ajax({
    type : 'POST',
    url: 'systemSettings',
    data: JSON.stringify(gSet),
    success: function () {
      hideAllPages();
      document.getElementById("info_text").innerHTML = "<br><div class='row'><div class='col'><p><label>Настройки успешно сохранены</label></p></div></div>";
      showPage("pageAreaInfo"); 
    }
  });
}

function loadSystemSettings() {
  $.get("systemSettings", {}, function(result) {
    //alert(result);
    var data = JSON.parse(result);
    document.getElementById("valSsid").value = data.valSsid;
    document.getElementById("valPass").value = data.valPass;
    document.getElementById("valAPName").value = data.valAPName;
    showSystemSettingsPage();
  });
}


function loadMainStatus() {
  $.get("mainData", {}, function(result) {
    var data = JSON.parse(result, function(key, val){
      if (val == "true") {
        return "включен";
      } else if (val == "false") {
        return "выключен";
      } else {
        return val;
      }
    });
    document.getElementById("temp_val").innerHTML = data.temp_val;
    document.getElementById("hum_val").innerHTML = data.hum_val;
    document.getElementById("light_state").innerHTML = data.light_state;
    document.getElementById("pump_state").innerHTML = data.pump_state;
    document.getElementById("fan_state").innerHTML = data.fan_state;
    document.getElementById("chiller_state").innerHTML = data.chiller_state;
    document.getElementById("date_time").innerHTML = data.date_time;
  });
}