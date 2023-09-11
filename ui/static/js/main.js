var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

function show() {
    const image = document.getElementById('image');
    const btn = document.getElementById('btnID');
    
    image.style.display = "block";
    btn.style.display = "none";
}

function openPopup(imagePath) {
    const windowFeatures = 'toolbar=no, location=no, status=no, menubar=no, scrollbars=yes, resizable=yes, width=1090, height=550, top=25, left=120';
    window.open(imagePath, 'targetWindow', windowFeatures);
}

const image = document.getElementById('image');
image.addEventListener('click', function () {
    const imagePath = this.getAttribute('src');
    openPopup(imagePath);
});
