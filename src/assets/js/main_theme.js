// script.js (ES module)
// Работает с three.js (ESM), SimplexNoise и IntersectionObserver
// Для стабильной работы запускай через локальный http сервер (file:// может блокировать модули)

import * as THREE from 'https://unpkg.com/three@0.152.2/build/three.module.js';
import { OrbitControls } from 'https://unpkg.com/three@0.152.2/examples/jsm/controls/OrbitControls.js';

//
// UI: reveal on scroll
//

class SimplexNoise {
    constructor(randomOrSeed) {
        let random = Math.random;
        if (typeof randomOrSeed === 'function') random = randomOrSeed;
        else if (typeof randomOrSeed === 'string' || typeof randomOrSeed === 'number') {
            // simple seeded RNG (xorshift32)
            let seed = 0x9E3779B9 ^ Number(String(randomOrSeed));
            random = function() {
                seed = (seed ^ (seed << 13)) >>> 0;
                seed = (seed ^ (seed >>> 17)) >>> 0;
                seed = (seed ^ (seed << 5)) >>> 0;
                return ((seed >>> 0) / 4294967295);
            };
        }
        this.p = new Uint8Array(256);
        for (let i = 0; i < 256; i++) this.p[i] = i;
        for (let i = 255; i > 0; i--) {
            const r = Math.floor(random() * (i + 1));
            const tmp = this.p[i];
            this.p[i] = this.p[r];
            this.p[r] = tmp;
        }
        this.perm = new Uint8Array(512);
        this.permGradIndex3 = new Uint8Array(512);
        for (let i = 0; i < 512; i++) {
            this.perm[i] = this.p[i & 255];
            this.permGradIndex3[i] = (this.perm[i] % 12) * 3;
        }

        this.grad3 = new Float32Array([
            1,1,0, -1,1,0, 1,-1,0, -1,-1,0,
            1,0,1, -1,0,1, 1,0,-1, -1,0,-1,
            0,1,1, 0,-1,1, 0,1,-1, 0,-1,-1
        ]);
    }

    // 3D simplex noise
    noise3D(xin, yin, zin) {
        const grad3 = this.grad3;
        const perm = this.perm;

        const F3 = 1 / 3;
        const G3 = 1 / 6;

        let n0 = 0, n1 = 0, n2 = 0, n3 = 0;

        const s = (xin + yin + zin) * F3;
        const i = Math.floor(xin + s);
        const j = Math.floor(yin + s);
        const k = Math.floor(zin + s);

        const t = (i + j + k) * G3;
        const X0 = i - t;
        const Y0 = j - t;
        const Z0 = k - t;

        const x0 = xin - X0;
        const y0 = yin - Y0;
        const z0 = zin - Z0;

        // rank for simplex corner ordering
        let i1, j1, k1;
        let i2, j2, k2;

        if (x0 >= y0) {
            if (y0 >= z0)      { i1=1; j1=0; k1=0;  i2=1; j2=1; k2=0; }
            else if (x0 >= z0) { i1=1; j1=0; k1=0;  i2=1; j2=0; k2=1; }
            else               { i1=0; j1=0; k1=1;  i2=1; j2=0; k2=1; }
        } else {
            if (y0 < z0)       { i1=0; j1=0; k1=1;  i2=0; j2=1; k2=1; }
            else if (x0 < z0)  { i1=0; j1=1; k1=0;  i2=0; j2=1; k2=1; }
            else               { i1=0; j1=1; k1=0;  i2=1; j2=1; k2=0; }
        }

        const x1 = x0 - i1 + G3;
        const y1 = y0 - j1 + G3;
        const z1 = z0 - k1 + G3;
        const x2 = x0 - i2 + 2*G3;
        const y2 = y0 - j2 + 2*G3;
        const z2 = z0 - k2 + 2*G3;
        const x3 = x0 - 1 + 3*G3;
        const y3 = y0 - 1 + 3*G3;
        const z3 = z0 - 1 + 3*G3;

        const ii = i & 255;
        const jj = j & 255;
        const kk = k & 255;

        // contributions
        let t0 = 0.6 - x0*x0 - y0*y0 - z0*z0;
        if (t0 > 0) {
            const gi0 = (perm[ii + perm[jj + perm[kk]]] % 12) * 3;
            t0 *= t0;
            n0 = t0 * t0 * (grad3[gi0]*x0 + grad3[gi0+1]*y0 + grad3[gi0+2]*z0);
        }

        let t1 = 0.6 - x1*x1 - y1*y1 - z1*z1;
        if (t1 > 0) {
            const gi1 = (perm[ii + i1 + perm[jj + j1 + perm[kk + k1]]] % 12) * 3;
            t1 *= t1;
            n1 = t1 * t1 * (grad3[gi1]*x1 + grad3[gi1+1]*y1 + grad3[gi1+2]*z1);
        }

        let t2 = 0.6 - x2*x2 - y2*y2 - z2*z2;
        if (t2 > 0) {
            const gi2 = (perm[ii + i2 + perm[jj + j2 + perm[kk + k2]]] % 12) * 3;
            t2 *= t2;
            n2 = t2 * t2 * (grad3[gi2]*x2 + grad3[gi2+1]*y2 + grad3[gi2+2]*z2);
        }

        let t3 = 0.6 - x3*x3 - y3*y3 - z3*z3;
        if (t3 > 0) {
            const gi3 = (perm[ii + 1 + perm[jj + 1 + perm[kk + 1]]] % 12) * 3;
            t3 *= t3;
            n3 = t3 * t3 * (grad3[gi3]*x3 + grad3[gi3+1]*y3 + grad3[gi3+2]*z3);
        }

        // scale to approx [-1,1]
        return 32 * (n0 + n1 + n2 + n3);
    }
}

/* -------------------------
   Основной код (three.js + glitch)
   ------------------------- */

// UI: reveal on scroll
(function initReveal(){
    const reveals = document.querySelectorAll('[data-reveal]');
    const obs = new IntersectionObserver((entries)=>{
        entries.forEach(e=>{
            if(e.isIntersecting){
                e.target.classList.add('show');
                obs.unobserve(e.target);
            }
        });
    }, {threshold: 0.15});
    reveals.forEach(r=>obs.observe(r));
})();

// Safety: z-index guard
document.addEventListener('DOMContentLoaded', ()=>{
    const containerHero = document.querySelector('.container-hero');
    if (containerHero) {
        containerHero.style.position = 'relative';
        containerHero.style.zIndex = '6';
    }
    const wrapEl = document.getElementById('three-wrap');
    if (wrapEl) {
        wrapEl.style.zIndex = '0';
        wrapEl.style.pointerEvents = 'none';
    }
    const gCanvas = document.getElementById('glitch-canvas');
    if (gCanvas) {
        gCanvas.style.pointerEvents = 'none';
        gCanvas.style.zIndex = '0';
    }
});

const wrap = document.getElementById('three-wrap');
const threeCanvas = document.getElementById('three-canvas');
const glitchCanvas = document.getElementById('glitch-canvas');

if (wrap && threeCanvas && glitchCanvas) {
    const DPR = Math.min(window.devicePixelRatio || 1, 2);

    const renderer = new THREE.WebGLRenderer({ canvas: threeCanvas, alpha: true, antialias: true });
    renderer.setPixelRatio(DPR);
    renderer.setSize(wrap.clientWidth, wrap.clientHeight, false);
    if ('outputColorSpace' in renderer) renderer.outputColorSpace = THREE.SRGBColorSpace;

    const scene = new THREE.Scene();
    const cam = new THREE.PerspectiveCamera(40, wrap.clientWidth / wrap.clientHeight, 0.1, 100);
    cam.position.set(0, 0.6, 2.6);

    const key = new THREE.DirectionalLight(0xa8ffca, 1.1); key.position.set(2,2,2); scene.add(key);
    const rim = new THREE.DirectionalLight(0x00ffaa, 0.35); rim.position.set(-2,1,-1); scene.add(rim);
    const amb = new THREE.AmbientLight(0x0b2b1e, 0.6); scene.add(amb);

    const baseGeom = new THREE.IcosahedronGeometry(1.05, 7);
    const geom = baseGeom.toNonIndexed();
    const posAttr = geom.attributes.position;
    const vertexCount = posAttr.count;

    const original = new Float32Array(vertexCount * 3);
    for (let i = 0; i < vertexCount * 3; i++) original[i] = posAttr.array[i];

    const material = new THREE.MeshStandardMaterial({
        color: 0x001a0f,
        emissive: 0x00ff9a,
        emissiveIntensity: 0.45,
        metalness: 0.15,
        roughness: 0.45,
        transparent: true,
        opacity: 0.98,
        side: THREE.DoubleSide
    });

    const mesh = new THREE.Mesh(geom, material);
    mesh.scale.setScalar(0.95);
    const group = new THREE.Group();
    group.add(mesh);
    scene.add(group);

    const controls = new OrbitControls(cam, renderer.domElement);
    controls.enablePan = false; controls.enableZoom = false; controls.enableRotate = false;

    // Используем встроенный SimplexNoise
    const noise = new SimplexNoise();

    const glitchCtx = glitchCanvas.getContext('2d');

    function onResize() {
        const w = wrap.clientWidth;
        const h = wrap.clientHeight;
        renderer.setSize(w, h, false);
        cam.aspect = w / h; cam.updateProjectionMatrix();

        glitchCanvas.width = Math.floor(w * DPR);
        glitchCanvas.height = Math.floor(h * DPR);
        glitchCanvas.style.width = `${w}px`;
        glitchCanvas.style.height = `${h}px`;
        glitchCtx.setTransform(DPR, 0, 0, DPR, 0, 0);
    }
    window.addEventListener('resize', onResize, { passive: true });
    onResize();

    // vertex update
    let t = 0;
    const speed = 0.45;
    const amplitude = 0.26;
    function updateVertices(time) {
        t = time * 0.0008 * speed;
        const arr = posAttr.array;
        for (let i = 0; i < vertexCount; i++) {
            const i3 = i * 3;
            const ox = original[i3], oy = original[i3 + 1], oz = original[i3 + 2];
            const n = noise.noise3D(ox * 1.6 + t * 0.8, oy * 1.6 + t * 0.9, oz * 1.6 + t * 1.2);
            const n2 = noise.noise3D(ox * 4.0 + t * 0.6, oy * 4.0 - t * 0.5, oz * 4.0 + t * 0.7) * 0.45;
            const disp = (n * 0.65 + n2 * 0.35) * amplitude;
            const len = Math.sqrt(ox*ox + oy*oy + oz*oz) || 1;
            arr[i3]     = ox + (ox/len) * disp;
            arr[i3 + 1] = oy + (oy/len) * disp;
            arr[i3 + 2] = oz + (oz/len) * disp;
        }
        posAttr.needsUpdate = true;
        geom.computeVertexNormals();
    }

    let lastGlitch = { time: 0 };
    const glitchIntervalBase = 1200;

    function renderGlitch() {
        const source = renderer.domElement;
        const cw = glitchCanvas.width / DPR;
        const ch = glitchCanvas.height / DPR;

        glitchCtx.clearRect(0, 0, cw, ch);

        // scanlines
        glitchCtx.globalAlpha = 0.04;
        glitchCtx.fillStyle = '#000';
        const lines = 120;
        for (let i = 0; i < lines; i++) {
            const y = (i / lines) * ch;
            glitchCtx.fillRect(0, y, cw, 1 * (1 / DPR));
        }
        glitchCtx.globalAlpha = 1;

        // grain
        const grainAmount = 0.02;
        const grainCount = Math.floor(cw * ch * grainAmount * 0.002);
        glitchCtx.fillStyle = 'rgba(255,255,255,0.02)';
        for (let i=0; i<grainCount; i++){
            const x = Math.random() * cw;
            const y = Math.random() * ch;
            glitchCtx.fillRect(x, y, 1, 1);
        }

        glitchCtx.save();
        glitchCtx.globalCompositeOperation = 'lighter';
        glitchCtx.globalAlpha = 0.22;
        glitchCtx.drawImage(source, 0, 0, cw, ch);
        glitchCtx.globalAlpha = 1;
        glitchCtx.restore();

        const now = performance.now();
        if (now - lastGlitch.time > glitchIntervalBase + Math.random()*2000) {
            lastGlitch.time = now;
            const slices = 6 + Math.floor(Math.random()*8);
            for (let i=0; i<slices; i++){
                const sy = Math.random() * ch;
                const sh = 6 + Math.random() * (ch * 0.12);
                const sx = Math.random() * cw;
                const w = cw * (0.18 + Math.random()*0.4);
                try {
                    const data = glitchCtx.getImageData(0, sy, cw, sh);
                    const dx = (Math.random()*2 - 1) * (cw * 0.06);
                    glitchCtx.putImageData(data, dx, sy);
                    glitchCtx.globalCompositeOperation = 'screen';
                    glitchCtx.fillStyle = `rgba(0,255,160,${0.06 + Math.random()*0.08})`;
                    glitchCtx.fillRect(sx, sy, w, sh);
                    glitchCtx.globalCompositeOperation = 'source-over';
                } catch(e) {
                    // ignore getImageData errors (CORS on canvas) — shouldn't happen for local content
                }
            }
        }
    }

    const clock = new THREE.Clock();
    function animate() {
        const time = clock.getElapsedTime();
        updateVertices(time * 1000);
        group.rotation.y = Math.sin(time * 0.2) * 0.25;
        group.rotation.x = Math.sin(time * 0.07) * 0.06;
        renderer.render(scene, cam);
        renderGlitch();
        requestAnimationFrame(animate);
    }
    requestAnimationFrame(animate);
} // end if

// demo login button
const loginBtn = document.getElementById('loginBtn');
if (loginBtn) {
    loginBtn.addEventListener('click', (e)=>{
        e.preventDefault();
        alert('Тут будет форма входа (OAuth) — могу добавить, Люцифер.');
    });
}