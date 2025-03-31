import { GraphQLClient, gql } from 'graphql-request';

const endpoint = 'https://gateway.thegraph.com/api/subgraphs/id/9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk';

const client = new GraphQLClient(endpoint, {
  headers: {
    Authorization: 'Bearer f80997c7afbb7cb5e69cbffbc9c583bf',
  },
});

const query = gql`
{
  factories(first: 5) {
    id
    poolCount
    txCount
    totalVolumeUSD
  }
  bundles(first: 5) {
    id
    nativePriceUSD
  }
}
`;

async function fetchData() {
  try {
    const data = await client.request(query);
    console.log(data);
  } catch (error) {
    console.error('Error fetching data:', error);
  }
}

fetchData();