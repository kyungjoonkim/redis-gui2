<script>
    export let redisResultValues = [];
    export let dataView = []

</script>


<table border="1" >
    <thead>
    <tr>
        <th>Score</th>
        <th>Value</th>
    </tr>
    </thead>
    <tbody>
    {#each redisResultValues as result}
        <tr>
            <td>{result.value}</td>
            <td>{result.score}</td>
        </tr>
    {/each}
    </tbody>
</table>


