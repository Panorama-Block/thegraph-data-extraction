import { GraphQLClient, gql } from 'graphql-request';
import dotenv from 'dotenv';

dotenv.config();

const endpoints = JSON.parse(process.env.ENDPOINTS_JSON);
const authToken = process.env.GRAPHQL_AUTH_TOKEN;

// Definição das queries
const queries = [
  gql`{
    transactions(first: 5) {
      id
      logIndex
      event
      from
    }
    tokens(first: 5) {
      id
      transaction { id }
      vault { id }
      activationBlock
    }
    _meta {
      deployment
      hasIndexingErrors
      block { hash number parentHash timestamp }
    }
  }`,
  gql`{
    factories(first: 5) {
      id
      poolCount
      txCount
      totalVolumeUSD
      owner
    }
  }`,
  gql`{
    factories(first: 5) {
      id
      poolCount
      txCount
      totalVolumeUSD
      owner
      totalFeesUSD
      totalFeesETH
    }
  }`
];

async function fetchData(endpoint, query) {
  const client = new GraphQLClient(`https://gateway.thegraph.com/api/subgraphs/id/${endpoint}`, {
    headers: {
      Authorization: `Bearer ${authToken}`,
    },
  });

  try {
    const data = await client.request(query);
    console.log(`\nData from ${endpoint}:`, data);
  } catch (error) {
    console.error(`Error fetching data from ${endpoint}:`, error);
  }
}

for (let i = 0; i < endpoints.length; i++) {
  fetchData(endpoints[i], queries[i]);
}