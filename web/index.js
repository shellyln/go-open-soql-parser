
const myCodeMirrorIn = CodeMirror.fromTextArea(document.querySelector('#soql'), {mode:'sql', theme: 'dracula', lineNumbers: true});
const myCodeMirrorOut = CodeMirror.fromTextArea(document.querySelector('#result-area'), {mode:'javascript', theme: 'dracula', lineNumbers: true, lineWrapping: true});

function execParse(event) {
    event.preventDefault();
    try {
        let result = parseSoql(myCodeMirrorIn.getDoc().getValue());
        myCodeMirrorOut.getDoc().setValue(result);
    } catch(e) {
        rebootGoApplication();
        myCodeMirrorOut.getDoc().setValue(e.message ?? e);
    }
}

async function copyToClipboard(event) {
    event.preventDefault();
    const text = myCodeMirrorOut.getDoc().getValue();
    try {
        await navigator.clipboard.writeText(text);
        const el = document.querySelector('#copied');
        el.style.display = 'inline-block';
        setTimeout(() => {
            el.style.display = 'none';
        }, 1200);
    } catch (e) {
        //
    }
}
