function highlightActiveNavLink() {
    const navLinks = document.querySelectorAll(".nav a");
    const currentPath = window.location.pathname;

    navLinks.forEach((link) => {
        if (link.getAttribute("href") === currentPath) {
            link.classList.add("live");
        }
    });
}

function openPopup(imagePath) {
    const windowFeatures = "toolbar=no, location=no, status=no, menubar=no, scrollbars=yes, resizable=yes, width=1090, height=550, top=25, left=120";
    window.open(imagePath, "targetWindow", windowFeatures);
}

function bindForumImageInteractions() {
    const image = document.getElementById("image");
    const revealButton = document.getElementById("btnID");

    if (!image || !revealButton) {
        return;
    }

    revealButton.addEventListener("click", () => {
        image.style.display = "block";
        revealButton.style.display = "none";
    });

    image.addEventListener("click", () => {
        const imagePath = image.getAttribute("src");
        if (imagePath) {
            openPopup(imagePath);
        }
    });
}

document.addEventListener("DOMContentLoaded", () => {
    highlightActiveNavLink();
    bindForumImageInteractions();
});