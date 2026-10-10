document.getElementById("longurl").onsubmit = async function(event) {
    event.preventDefault();
    try {
        let longUrl = document.getElementById("lurl").value;
        const response = await fetch("http://localhost:8090/",{
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                url: longUrl
            })
        });
        if(!response.ok) {
            alert("something went wrong");
            return;
        } else {
            const data = await response.json();
            const shortUrl = data.short_url;
            document.getElementById("surl").value = shortUrl;
        } 
    } catch (error) {
       alert("Request failed. Please try again");
    } 
};

function copyUrl(){
    const shortUrl = document.getElementById("surl").value;
    navigator.clipboard.writeText(shortUrl)
    .then(() => {
        document.getElementById("copy-status").textContent = "Copied to clipboard!";
    })
    .catch(err => {
        document.getElementById("copy-status").textContent = "Failed to copy";
    });
}

document.getElementById("delete").onclick = async function(event) {
    try {
        const shortUrl = document.getElementById("surl").value;
        const key = new URL(shortUrl).pathname.slice(1);
        const response = await fetch("http://localhost:8090/"+key, {
            method: "DELETE",
            headers: {
                "Content-Type": "application-json"
            }, 
        });
        if (!response.ok) {
           alert("something went wrong");
           return
        }
        document.getElementById("surl").value = "";
    } catch (error) {
       alert("Request failed. Please try again");
    }
}