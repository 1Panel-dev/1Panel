window.onload = function () {
    setTimeout(function () {
        showSuccess();
        verifySucc();
    }, 5000);

    function showSuccess() {
        document.getElementById("loadingText").style.display = "none";
        document.getElementById("loadingSuccess").style.display = "block";
        document.querySelector(".loadingSpinner").style.display = "none";
    }

    function verifySucc() {
        let xhr = new XMLHttpRequest();
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4 && xhr.status === 200) {
                window.location.reload();
            }
        };
        const requestUrl = "%s-%s-%s";
        xhr.open("GET", requestUrl, true);
        xhr.send();
    }
}