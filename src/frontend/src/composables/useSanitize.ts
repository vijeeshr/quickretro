import DOMPurify from 'dompurify';

export default () => ({
    sanitize: (dirty: string) => DOMPurify.sanitize(dirty, {
        ALLOWED_TAGS: ['b', 'i', 'u', 'em', 'strong', 'br'],
        ALLOWED_ATTR: []
    })
})