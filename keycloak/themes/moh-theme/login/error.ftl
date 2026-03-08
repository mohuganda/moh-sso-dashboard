<#import "template.ftl" as layout>

<@layout.registrationLayout>

<#if section = "form">

<div class="kc-error">
    <h2>Error</h2>

<p>${message.summary}</p>

<a href="${url.loginUrl}">Back to Login</a>

</div>

</#if>

[/@layout.registrationLayout](mailto:/@layout.registrationLayout)
