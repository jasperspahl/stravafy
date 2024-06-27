class LocalTimestamp extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({ mode: 'open' });
	}

	connectedCallback() {
		const date = this.getAttribute('date');
		this.shadowRoot.innerHTML = new Date(date).toLocaleString();
	}
}

customElements.define('local-timestamp', LocalTimestamp);
