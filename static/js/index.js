var time = 0;
var timer = 0;
function DecTime() { 
    if (time > 0) {
        time--;
        document.getElementById("clockdiv").innerHTML = 'До возвращения в автоматический режим: ' + time + ' сек';
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
    document.getElementById("clockdiv").innerHTML = 'До возвращения в автоматический режим: ' + time + ' сек';
    timer = setInterval(function() { 
        DecTime(); 
    }, 1000); 
} 

function sendRequestToMain(request) {
          $.get("/mainData/"+request, {}, function(result) {
            //alert(result);
            restartCountdown(result);
          });
}

function initApp() {
  setInterval(function() {
    loadMainStatus();
  }, 1000);
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

function saveGrowingSettings() {
  var gSet = {
    valLightOnTime: document.getElementById("valLightOnTime").value,
    valLightOffTime: document.getElementById("valLightOffTime").value,
    valPumpOnTime: document.getElementById("valPumpOnTime").value,
    valPumpPauseTime: document.getElementById("valPumpPauseTime").value,
    valFanOnThreshold: document.getElementById("valFanOnThreshold").value
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
  if (gSet.valFanOnThreshold == "") {
    gSet.valFanOnThreshold = document.getElementById("valFanOnThreshold").placeholder;
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
    document.getElementById("valFanOnThreshold").value = data.valFanOnThreshold;
    document.getElementById("valPumpPauseTime").value = data.valPumpPauseTime;
    showGrowingSettingsPage();
  });
  
}

function saveSystemSettings() {
  var gSet = {
    valSsid: document.getElementById("valSsid").value,
    valPass: document.getElementById("valPass").value,
    valAPName: document.getElementById("valAPName").value
  };
  if (gSet.valSsid == "") {
    gSet.valSsid = document.getElementById("valSsid").placeholder;
  }
  if (gSet.valPass == "") {
    gSet.valPass = document.getElementById("valPass").placeholder;
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
    var data = JSON.parse(result);
    document.getElementById("temp_val").value = data.temp_val;
    document.getElementById("hum_val").value = data.hum_val;
    document.getElementById("light_state").value = data.light_state;
    document.getElementById("pump_state").value = data.pump_state;
    document.getElementById("fan_state").value = data.fan_state;
  });
}