document.getElementById("longurl").onsubmit = async function(event) {
    event.preventDefault();
    try {
        let longUrl = document.getElementById("lurl").value;
        const response = await fetch("https://url-shortener-qj2v.onrender.com/",{
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

document.getElementById("copy").onclick = function() {
    const shortUrl = document.getElementById("surl").value;
    navigator.clipboard.writeText(shortUrl)
    .then(() => {
        document.getElementById("copy-status").textContent = "Copied to clipboard!";
    })
    .catch(err => {
        document.getElementById("copy-status").textContent = "Failed to copy";
    });
}