function Check() {
    document.getElementById('left').classList.toggle("hidden");
}

// function ReadId(e) { 
//     alert(e.id);
// }

function Drop(id) {
    const scrollToBottom = (id) => {
        const element = document.getElementById(id);
        element.scrollTop = element.scrollHeight;
    }
}
