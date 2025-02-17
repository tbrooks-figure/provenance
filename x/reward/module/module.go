package reward

import (
	"context"
	"encoding/json"
	"math/rand"

	rewardModule "github.com/provenance-io/provenance/x/reward"
	cli "github.com/provenance-io/provenance/x/reward/client/cli"
	"github.com/provenance-io/provenance/x/reward/keeper"
	simulation "github.com/provenance-io/provenance/x/reward/simulation"
	"github.com/provenance-io/provenance/x/reward/types"

	cerrs "cosmossdk.io/errors"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the reward module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the reward module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// RegisterLegacyAminoCodec registers the reward module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
}

// RegisterInterfaces registers the reward module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the reward
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the reward module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return cerrs.Wrapf(err, "failed to unmarshal %q genesis state", types.ModuleName)
	}

	return data.Validate()
}

// RegisterRESTRoutes registers the REST routes for the reward module.
// Deprecated: RegisterRESTRoutes is deprecated.
func (AppModuleBasic) RegisterRESTRoutes(_ sdkclient.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the reward module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetQueryCmd returns the cli query commands for the reward module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// GetTxCmd returns the transaction commands for the reward module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper        keeper.Keeper
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, accountKeeper authkeeper.AccountKeeper, bankKeeper bankkeeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

// GenerateGenesisState creates a randomized GenState of the rewards module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns content functions used to simulate governance proposals.
func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	// currently no gov proposals exist
	return nil
}

// RandomizedParams returns randomized module parameters for param change proposals.
func (am AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	// currently no module params exist
	return nil
}

// RegisterStoreDecoder registers a func to decode each module's defined types from their corresponding store key
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[keeper.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns simulation operations (i.e msgs) with their respective weight
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, am.keeper, am.accountKeeper, am.bankKeeper,
	)
}

// Name returns the reward module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the reward module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, rewardModule.NewHandler(am.keeper))
}

// QuerierRoute returns the route we respond to for abci queries
func (AppModule) QuerierRoute() string { return "" }

// LegacyQuerierHandler returns the reward module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs genesis initialization for the reward module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the reward
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock is the `BeginBlocker` function run at the beginning of each block to
// process rewards module updates.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	rewardModule.BeginBlocker(ctx, am.keeper)
}

// EndBlock The `EndBlocker` abci call is ran at the end of each block. The `EventManager` is monitored
// and `Qualifying Actions` are deduced from newly created events and prior internal state.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	rewardModule.EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}
